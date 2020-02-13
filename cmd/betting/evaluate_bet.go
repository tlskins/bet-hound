package betting

import (
	"fmt"
	"reflect"
	"time"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

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
		winner = *bet.Recipient
		loser = &bet.Proposer
	}
	bet.BetResult = &t.BetResult{
		Winner:    winner,
		Loser:     *loser,
		Response:  bet.ResultString(),
		DecidedAt: time.Now(),
	}
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
		return evaluatePlayerExpression(p, *g)
	} else if t, ok := e.(t.TeamExpression); ok {
		return evaluateTeamExpression(t, *g)
	}
	return nil, fmt.Errorf("Unable to evaluate expression type.")
}

func evaluatePlayerExpression(e t.PlayerExpression, g t.GameAndLog) (t.Expression, error) {
	if err := e.Valid(); err != nil {
		return nil, err
	}

	gameLog := g.GetGameLog()
	playerLog := gameLog.PlayerLogFor(e.Player.Fk)
	// inactive player
	if reflect.ValueOf(playerLog).IsNil() {
		zero := 0.0
		e.Value = &zero
	} else {
		e.Value = playerLog.EvaluateMetric(e.Metric.Field)
	}
	var expr t.Expression = e
	return expr, nil
}

func evaluateTeamExpression(e t.TeamExpression, g t.GameAndLog) (t.Expression, error) {
	if err := e.Valid(); err != nil {
		return nil, err
	}

	gameLog := g.GetGameLog()
	playerLog := gameLog.PlayerLogFor(e.Team.Fk)
	e.Value = playerLog.EvaluateMetric(e.Metric.Field)
	var expr t.Expression = e
	return expr, nil
}

func getExpressionGameAndLog(expr t.Expression) (*t.GameAndLog, error) {
	gm := expr.GetGame()
	gmAndLog, err := db.FindGameAndLogById(gm.Id, gm.LeagueId)
	if err != nil || gmAndLog == nil {
		errResp := err
		if errResp == nil {
			errResp = fmt.Errorf("Game and log invalid %s", gm.Id)
		}
		return nil, errResp
	}
	return gmAndLog, err
}
