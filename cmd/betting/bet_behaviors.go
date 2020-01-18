package betting

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/satori/go.uuid"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

func AcceptBet(user *t.User, betId string, accept bool) (bool, error) {
	bet, err := db.FindBetById(betId)
	if err != nil {
		return false, err
	} else if bet.BetStatus.String() != "Pending Approval" {
		return false, fmt.Errorf("Cannot accept a bet with status: %s", bet.BetStatus.String())
	} else if bet.Recipient.Id != user.Id && bet.Proposer.Id != user.Id {
		return false, fmt.Errorf("You are not involved with this bet")
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
		return false, err
	}
	return true, nil
}

func CreateBet(proposer *t.User, changes t.BetChanges) (bet *t.Bet, err error) {
	now := time.Now()
	rand.Seed(now.UnixNano())
	recipient, err := db.FindUserById(changes.RecipientId)
	if err != nil {
		return nil, err
	}
	pReplyFk := "-1"
	bet = &t.Bet{
		Id:              uuid.NewV4().String(),
		BetStatus:       t.BetStatusFromString("Pending Approval"),
		ProposerReplyFk: &pReplyFk,
		CreatedAt:       &now,
		Proposer:        *proposer,
		Recipient:       *recipient,
	}

	// get bet map lookups
	settings, err := db.GetLeagueSettings("nfl")
	if err != nil {
		return nil, err
	}
	opMap := settings.BetEquationsMap()
	metricMap := settings.PlayerBetsMap()

	// create equations
	for _, eqChg := range changes.EquationsChanges {
		eq := &t.Equation{Id: rand.Intn(9999999)}
		// create operator
		if eqChg.OperatorId != nil {
			eq.Operator = opMap[*eqChg.OperatorId]
		}
		// create expression
		for _, exprChg := range eqChg.ExpressionChanges {
			expr := &t.PlayerExpression{Id: rand.Intn(9999999)}
			if exprChg.IsLeft != nil {
				expr.IsLeft = *exprChg.IsLeft
			}
			// add player
			if exprChg.PlayerFk != nil {
				expr.Player, _ = db.FindPlayer(*exprChg.PlayerFk)
			}
			// add game
			if exprChg.GameFk != nil {
				expr.Game, _ = db.FindCurrentGame(settings, *exprChg.GameFk)
			}
			// add metric
			if exprChg.MetricId != nil {
				expr.Metric = metricMap[*exprChg.MetricId]
			}
			eq.Expressions = append(eq.Expressions, expr)
		}
		bet.Equations = append(bet.Equations, eq)
	}

	// validate and upsert
	bet.PostProcess()
	if err = bet.Valid(); err != nil {
		return nil, err
	}
	if err = db.UpsertBet(bet); err != nil {
		return nil, err
	}
	if bet.Recipient.TwitterUser != nil {
		if _, err = TweetBetProposal(bet); err != nil {
			return bet, err
		}
	}
	return
}

func EvaluateBet(b *t.Bet, g *t.Game) (*t.Bet, error) {
	bet := *b
	betComplete := true
	pWins := true
	for _, eq := range bet.Equations {
		eqComplete := true
		// evaluate expressions involving this game
		for _, expr := range eq.Expressions {
			if expr.Game.Id == g.Id {
				e, err := EvaluateExpression(expr, g)
				if err != nil {
					return nil, err
				}
				expr = e
			} else if expr.Value == nil {
				betComplete = false
				eqComplete = false
			}
		}
		// evaluate complete equations
		if eqComplete {
			e, err := EvaluateEquation(eq)
			if err != nil {
				return nil, err
			}
			eq = e
			if !*eq.Result {
				pWins = false
			}
		}
	}
	if betComplete {
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
	}
	return &bet, nil
}

func EvaluateEquation(e *t.Equation) (*t.Equation, error) {
	eq := *e
	left, right := 0.0, 0.0
	for _, expr := range eq.Expressions {
		if expr.IsLeft {
			left += *expr.Value
		} else {
			right += *expr.Value
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
		return nil, fmt.Errorf("Unsupported bet operator")
	}
	return &eq, nil
}

func EvaluateExpression(e *t.PlayerExpression, g *t.Game) (expr *t.PlayerExpression, err error) {
	if err = e.Valid(); err != nil {
		return nil, err
	}
	if e.Game.Id != g.Id {
		return nil, fmt.Errorf("Game and expression dont match")
	}
	if g.GameLog == nil {
		return nil, fmt.Errorf("Game logs missing")
	}

	log := g.GameLog.PlayerLogs[e.Player.Id]
	if log == nil {
		zero := 0.0
		e.Value = &zero
		return e, nil
	} else {
		r := reflect.ValueOf(log)
		r = r.Elem()
		v := r.FieldByName(e.Metric.Field)
		value := v.Float()
		e.Value = &value
		return e, nil
	}
}

// func CalcBetResult(bet *t.Bet) (err error) {
// 	fmt.Println("calc bet result ", bet.Id, bet.String())

// 	responses := []string{}
// 	proposerWins := true
// 	for _, eq := range bet.Equations {
// 		eqResult, err := calcEquationResult(eq, &games)
// 		if err != nil {
// 			return err
// 		}
// 		responses = append(responses, fmt.Sprintf("%s (%t)", eq.ResultDescription(), *eqResult))
// 		proposerWins = proposerWins && *eqResult
// 	}

// 	var wUsr, lUsr t.User
// 	if proposerWins {
// 		wUsr = bet.Proposer
// 		lUsr = bet.Recipient
// 	} else {
// 		wUsr = bet.Recipient
// 		lUsr = bet.Proposer
// 	}
// 	bet.BetResult = &t.BetResult{
// 		Winner: wUsr,
// 		Loser:  lUsr,
// 		Response: fmt.Sprintf("Congrats @%s you beat @%s! '%s'",
// 			wUsr.ScreenName,
// 			lUsr.ScreenName,
// 			strings.Join(responses, ", "),
// 		),
// 		DecidedAt: time.Now(),
// 	}
// 	bet.BetStatus = t.BetStatusFromString("Final")
// 	return nil
// }

// func findGameByFk(games *[]*t.Game, teamFk string) *t.Game {
// 	for _, g := range *games {
// 		if g.AwayTeamFk == teamFk || g.HomeTeamFk == teamFk {
// 			return g
// 		}
// 	}
// 	return nil
// }

// // helpers

// func calcEquationResult(eq *t.Equation, games *[]*t.Game) (*bool, error) {
// 	// Process each expression
// 	allExprs := [][]*t.PlayerExpression{eq.LeftExpressions, eq.RightExpressions}
// 	errs := []string{}
// 	for _, exprs := range allExprs {
// 		for _, expr := range exprs {
// 			err := calcExpressionResult(expr, games, eq.Metric)
// 			if err != nil {
// 				errs = append(errs, err.Error())
// 			}
// 		}
// 	}
// 	if len(errs) > 0 {
// 		return nil, fmt.Errorf(strings.Join(errs, ""))
// 	}

// 	// Record result
// 	lTtl := calcExpressionsTotal(&eq.LeftExpressions)
// 	rTtl := calcExpressionsTotal(&eq.RightExpressions)
// 	fixedMod := eq.Metric.FixedValueMod()
// 	var result bool
// 	if eq.Operator.Lemma == "more" {
// 		// Add fixed mods to equation
// 		if fixedMod != nil {
// 			*lTtl += *fixedMod
// 		}
// 		result = *lTtl > *rTtl
// 	} else if eq.Operator.Lemma == "less" || eq.Operator.Lemma == "few" {
// 		if fixedMod != nil {
// 			*lTtl -= *fixedMod
// 		}
// 		result = *lTtl < *rTtl
// 	}
// 	eq.Result = &result
// 	return &result, nil
// }

// func calcExpressionsTotal(expressions *[]*t.PlayerExpression) *float64 {
// 	total := 0.0
// 	for _, e := range *expressions {
// 		if e.Value == nil {
// 			return nil
// 		}
// 		total += *e.Value
// 	}

// 	return &total
// }

// func calcExpressionResult(expr *t.PlayerExpression, games *[]*t.Game, metric *t.Metric) (err error) {
// 	gm := findGameByFk(games, expr.Player.TeamFk)
// 	log := scraper.ScrapeGameLog(gm)
// 	value := calcPlayerGameValue(&log, expr.Player, metric)

// 	if value == nil {
// 		return fmt.Errorf("Unable to determine score for %s.", expr.Description())
// 	} else {
// 		expr.Value = value
// 		return nil
// 	}
// }
