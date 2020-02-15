package betting

import (
	"fmt"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

func CreateBet(proposer *t.User, newBet *t.NewBet) (bet *t.Bet, note *t.Notification, err error) {
	now := time.Now()
	rand.Seed(now.UnixNano())
	pReplyFk := "-1"
	bet = &t.Bet{
		Id:              uuid.NewV4().String(),
		LeagueId:        newBet.LeagueId,
		BetStatus:       t.BetStatusFromString("Pending Approval"),
		ProposerReplyFk: &pReplyFk,
		CreatedAt:       &now,
		Proposer:        *proposer.IndexUser(),
	}

	// add recipient
	if newBet.BetRecipient != nil {
		var recipient *t.IndexUser
		if rcp, err := db.FindOrCreateBetRecipient(newBet.BetRecipient); err != nil {
			return nil, nil, err
		} else {
			recipient = rcp.IndexUser()
		}
		if proposer.Id == recipient.Id {
			return nil, nil, fmt.Errorf("Can't make a bet with yourself!")
		}
		bet.Recipient = recipient
	}

	// get bet map lookups
	var betMapLookup map[int]*t.BetMap
	if betMaps, err := db.GetBetMaps(&newBet.LeagueId, nil); err != nil {
		return nil, nil, err
	} else {
		betMapLookup = make(map[int]*t.BetMap)
		for _, betMap := range betMaps {
			betMapLookup[betMap.Id] = betMap
		}
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
			if expr, err := createExpression(newExpr, &betMapLookup); err != nil {
				return nil, nil, err
			} else {
				eq.Expressions = append(eq.Expressions, expr)
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
	if bet.Recipient != nil && bet.Recipient.TwitterUser != nil {
		if _, err = TweetBetProposal(bet); err != nil {
			return bet, nil, err
		}
	}
	note, err = db.SyncBetWithUsers("Create", bet)
	if err != nil {
		fmt.Println(err)
	}
	return bet, note, nil
}

// helpers

func createExpression(newExpr *t.NewExpression, betMapLookup *map[int]*t.BetMap) (expr t.Expression, err error) {
	var player *t.Player
	var game *t.Game
	var team *t.Team
	var metric *t.BetMap

	// lookup components
	if newExpr.PlayerId != nil {
		if player, err = db.FindPlayerById(*newExpr.PlayerId); err != nil {
			return
		}
	}
	if newExpr.GameId != nil {
		if game, err = db.FindGameById(*newExpr.GameId); err != nil {
			return
		}
	}
	if newExpr.TeamId != nil {
		if team, err = db.FindTeamById(*newExpr.TeamId); err != nil {
			return
		}
	}
	if newExpr.MetricId != nil {
		metric = (*betMapLookup)[*newExpr.MetricId]
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
			return
		}
		expr = px
	} else if team != nil {
		tx := t.TeamExpression{
			Id:     genPk(),
			Left:   newExpr.IsLeft,
			Team:   team,
			Game:   game,
			Metric: metric,
		}
		if err = tx.Valid(); err != nil {
			return
		}
		expr = tx
	} else if newExpr.Value != nil {
		sx := t.StaticExpression{
			Id:    genPk(),
			Value: newExpr.Value,
		}
		if err = sx.Valid(); err != nil {
			return
		}
		expr = sx
	}
	return
}

func genPk() int {
	return rand.Intn(9999999)
}
