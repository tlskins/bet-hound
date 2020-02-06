package betting

import (
	"fmt"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

func AcceptBet(user *t.User, betId string, accept bool) (*t.Bet, *t.Notification, error) {
	bet, err := db.FindBetById(betId)
	if err != nil {
		return nil, nil, err
	} else if bet.BetStatus.String() != "Pending Approval" {
		return nil, nil, fmt.Errorf("Cannot accept a bet with status: %s.", bet.BetStatus.String())
	} else if bet.Recipient.Id != user.Id && bet.Proposer.Id != user.Id {
		return nil, nil, fmt.Errorf("You are not involved with this bet.")
	}

	var status t.BetStatus
	if accept {
		dftFk := "-1"
		if bet.Proposer.Id == user.Id {
			bet.ProposerReplyFk = &dftFk
		} else if bet.Recipient.Id == user.Id {
			bet.RecipientReplyFk = &dftFk
		}

		if bet.ProposerReplyFk != nil && bet.RecipientReplyFk != nil {
			status = t.BetStatusFromString("Accepted")
		}
	} else {
		if bet.Proposer.Id == user.Id {
			status = t.BetStatusFromString("Withdrawn")
		} else if bet.Recipient.Id == user.Id {
			status = t.BetStatusFromString("Declined")
		}
	}
	bet.BetStatus = status
	if err = db.UpsertBet(bet); err != nil {
		return nil, nil, err
	}
	if _, err = TweetBetApproval(bet, nil); err != nil {
		return nil, nil, err
	}
	note, _ := db.SyncBetWithUsers("Update", bet)
	return bet, note, nil
}

func CreateBet(proposer *t.User, newBet *t.NewBet) (bet *t.Bet, note *t.Notification, err error) {
	now := time.Now()
	rand.Seed(now.UnixNano())
	recipient, err := db.FindOrCreateBetRecipient(&newBet.BetRecipient)
	if err != nil {
		return nil, nil, err
	}
	if proposer.Id == recipient.Id {
		return nil, nil, fmt.Errorf("Can't make a bet with yourself!")
	}
	pReplyFk := "-1"
	bet = &t.Bet{
		Id:              uuid.NewV4().String(),
		BetStatus:       t.BetStatusFromString("Pending Approval"),
		ProposerReplyFk: &pReplyFk,
		CreatedAt:       &now,
		Proposer:        *proposer.IndexUser(),
		Recipient:       *recipient.IndexUser(),
	}

	// get bet map lookups
	betMaps, err := db.GetBetMaps(&newBet.LeagueId, nil)
	if err != nil {
		return nil, nil, err
	}
	betMapLookup := make(map[int]*t.BetMap)
	for _, betMap := range betMaps {
		betMapLookup[betMap.Id] = betMap
	}

	// create equations
	for _, newEq := range newBet.NewEquations {
		eq := &t.Equation{Id: rand.Intn(9999999)}
		// create operator
		if newEq.OperatorId != nil {
			eq.Operator = betMapLookup[*newEq.OperatorId]
		}
		// create expression
		for _, newExpr := range newEq.NewExpressions {
			var player *t.Player
			var game *t.Game
			var team *t.Team
			var metric *t.BetMap

			if newExpr.PlayerId != nil {
				if player, err = db.FindPlayerById(*newExpr.PlayerId); err != nil {
					return nil, nil, err
				}
			}
			if newExpr.GameId != nil {
				if game, err = db.FindGameById(*newExpr.GameId); err != nil {
					return nil, nil, err
				}
			}
			if newExpr.TeamId != nil {
				if team, err = db.FindTeamById(*newExpr.TeamId); err != nil {
					return nil, nil, err
				}
			}
			if newExpr.MetricId != nil {
				metric = betMapLookup[*newExpr.MetricId]
			}

			if player != nil {
				px := t.PlayerExpression{
					Id:     genPk(),
					Left:   newExpr.IsLeft,
					Player: player,
					Game:   game,
					Metric: metric,
				}
				if err = px.Valid(); err != nil {
					return nil, nil, err
				}
				var ex t.Expression = px
				eq.Expressions = append(eq.Expressions, ex)
			} else if team != nil {
				tx := t.TeamExpression{
					Id:     genPk(),
					Left:   newExpr.IsLeft,
					Team:   team,
					Game:   game,
					Metric: metric,
				}
				if err = tx.Valid(); err != nil {
					return nil, nil, err
				}
				var ex t.Expression = tx
				eq.Expressions = append(eq.Expressions, ex)
			} else if newExpr.Value != nil {
				sx := t.StaticExpression{
					Id:    genPk(),
					Value: newExpr.Value,
				}
				if err = sx.Valid(); err != nil {
					return nil, nil, err
				}
				var ex t.Expression = sx
				eq.Expressions = append(eq.Expressions, ex)
			}
		}
		bet.Equations = append(bet.Equations, eq)
	}

	// validate and persistence
	bet.PostProcess()
	if err = bet.Valid(); err != nil {
		return nil, nil, err
	}
	if err = db.UpsertBet(bet); err != nil {
		return nil, nil, err
	}
	if bet.Recipient.TwitterUser != nil {
		if _, err = TweetBetProposal(bet); err != nil {
			return bet, nil, err
		}
	}
	note, _ = db.SyncBetWithUsers("Create", bet)
	return bet, note, nil
}

func EvaluateBet(b *t.Bet) (betResult *t.Bet, err error) {
	bet := *b
	pWins := true
	var gmAndLog *t.GameAndLog
	for _, eq := range bet.Equations {
		// evaluate expressions
		for i, expr := range eq.Expressions {
			if gmAndLog, err = getExpressionGameAndLog(expr); err != nil {
				return nil, err
			}
			if eq.Expressions[i], err = evaluateExpression(expr, gmAndLog); err != nil {
				return nil, err
			}
		}
		// evaluate complete equation
		if eq, err = evaluateEquation(eq); err != nil {
			return nil, err
		}
		if !*eq.Result {
			pWins = false
		}
	}

	// record result
	bet.BetStatus = t.BetStatusFromString("Final")
	winner := bet.Proposer
	loser := bet.Recipient
	if !pWins {
		winner = bet.Recipient
		loser = bet.Proposer
	}
	bet.BetResult = &t.BetResult{
		Winner:    winner,
		Loser:     loser,
		Response:  bet.ResultString(),
		DecidedAt: time.Now(),
	}

	// tweet result
	_, err = TweetBetResult(&bet)
	return &bet, err
}

// helpers

func evaluateEquation(e *t.Equation) (*t.Equation, error) {
	eq := *e
	left, right := 0.0, 0.0
	for _, expr := range eq.Expressions {
		result := expr.ResultValue()
		if result != nil {
			if expr.IsLeft() {
				left += *result
			} else {
				right += *result
			}
		}
	}

	t, f := true, false
	if eq.Operator.Field == "GreaterThan" {
		if left > right {
			eq.Result = &t
		} else {
			eq.Result = &f
		}
	} else if eq.Operator.Field == "LesserThan" {
		if left < right {
			eq.Result = &t
		} else {
			eq.Result = &f
		}
	} else if eq.Operator.Field == "Equal" {
		if left == right {
			eq.Result = &t
		} else {
			eq.Result = &f
		}
	} else {
		return nil, fmt.Errorf("Unsupported bet operator.")
	}
	return &eq, nil
}

func evaluateExpression(e t.Expression, g *t.GameAndLog) (expr t.Expression, err error) {
	if s, ok := e.(t.StaticExpression); ok {
		return s, nil
	} else if p, ok := e.(t.PlayerExpression); ok {
		return evaluatePlayerExpression(p, g)
	} else if t, ok := e.(t.TeamExpression); ok {
		return evaluateTeamExpression(t, g)
	}
	return nil, fmt.Errorf("Unable to evaluate expression type.")
}

func evaluatePlayerExpression(e t.PlayerExpression, g *t.GameAndLog) (t.Expression, error) {
	if err := e.Valid(); err != nil {
		return nil, err
	}
	gm := e.GetGame()
	if gm.Id != g.Id {
		return nil, fmt.Errorf("Game and expression dont match.")
	}
	if g.GameLog == nil {
		return nil, fmt.Errorf("Game logs missing.")
	}

	log := g.GameLog.PlayerLogs[e.Player.Id]
	e.Value = (*log).EvaluateMetric(e.Metric.Field)
	var expr t.Expression = e
	return expr, nil
}

func evaluateTeamExpression(e t.TeamExpression, g *t.GameAndLog) (t.Expression, error) {
	if err := e.Valid(); err != nil {
		return nil, err
	}
	gm := e.GetGame()
	if gm.Id != g.Id {
		return nil, fmt.Errorf("Game and expression dont match.")
	}
	if g.GameLog == nil {
		return nil, fmt.Errorf("Game logs missing.")
	}
	log := g.GameLog.TeamLogFor(e.Team.Fk)
	if log == nil {
		return nil, fmt.Errorf("Team logs missing.")
	} else {
		fmt.Println("team log ", *log)
	}

	e.Value = (*log).EvaluateMetric(e.Metric.Field)
	var expr t.Expression = e
	return expr, nil
}

func getExpressionGameAndLog(expr t.Expression) (*t.GameAndLog, error) {
	gm := expr.GetGame()
	gmAndLog, err := db.FindGameAndLogById(gm.Id)
	if err != nil || gmAndLog == nil || gmAndLog.GameLog == nil {
		errResp := err
		if errResp != nil {
			errResp = fmt.Errorf("Game and log invalid %s", gm.Id)
		}
		return nil, errResp
	}
	return gmAndLog, err
}

func genPk() int {
	return rand.Intn(9999999)
}
