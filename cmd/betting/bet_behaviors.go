package betting

import (
	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
	"fmt"
	"math/rand"
	// "strings"
	"time"

	"github.com/satori/go.uuid"
)

func UpdateBet(id string, changes t.BetChanges) (bet *t.Bet, err error) {
	bet, err = db.FindBetById(id)
	if err != nil {
		return nil, err
	}

	// get bet map lookups
	settings, err := db.GetLeagueSettings("nfl")
	if err != nil {
		return nil, err
	}
	opMap := settings.BetEquationsMap()
	metricMap := settings.PlayerBetsMap()

	// build equations map
	eqMap := map[int]*t.Equation{}
	eqIdxMap := map[int]*int{}
	for i, eq := range bet.Equations {
		eqMap[eq.Id] = eq
		eqIdxMap[eq.Id] = &i
	}

	// make equation changes
	for _, eqChg := range changes.EquationsChanges {
		eq := eqMap[eqChg.Id]
		if eq == nil {
			return nil, fmt.Errorf("equation not found")
		} else if eqChg.Delete != nil {
			// delete  equation
			idx := eqIdxMap[eqChg.Id]
			if idx != nil {
				copy(bet.Equations[*idx:], bet.Equations[*idx+1:])
				bet.Equations[len(bet.Equations)-1] = nil
				bet.Equations = bet.Equations[:len(bet.Equations)-1]
			}
		}
		// build expression map
		exprMap := map[int]*t.PlayerExpression{}
		exprIdxMap := map[int]*int{}
		for i, expr := range eq.Expressions {
			exprMap[expr.Id] = expr
			exprIdxMap[expr.Id] = &i
		}
		// operator changes
		if eqChg.OperatorId != nil {
			eq.Operator = opMap[*eqChg.OperatorId]
		}
		// expression changes
		for _, exprChg := range eqChg.ExpressionChanges {
			expr := exprMap[exprChg.Id]
			if expr == nil {
				return nil, fmt.Errorf("expression not found")
			} else if exprChg.Delete != nil {
				// delete  expression
				idx := exprIdxMap[exprChg.Id]
				if idx != nil {
					copy(bet.Equations[*idx:], bet.Equations[*idx+1:])
					bet.Equations[len(bet.Equations)-1] = nil
					bet.Equations = bet.Equations[:len(bet.Equations)-1]
				}
			}
			// change player
			if exprChg.PlayerFk != nil {
				if expr.Player, err = db.FindPlayer(*exprChg.PlayerFk); err != nil {
					return nil, err
				}
			}
			// change metric
			if exprChg.MetricId != nil {
				metric := metricMap[*exprChg.MetricId]
				if metric == nil {
					return nil, fmt.Errorf("not a valid metric")
				}
				expr.Metric = metric
			}
		}
	}

	err = db.UpsertBet(bet)
	return
}

func CreateBet(changes t.BetChanges) (bet *t.Bet, err error) {
	now := time.Now()
	bet = &t.Bet{
		Id:        uuid.NewV4().String(),
		CreatedAt: &now,
		Proposer:  t.User{Name: "Tim Lee", ScreenName: "steve_aoili"},
		Recipient: t.User{Name: "Christine Kettler", ScreenName: "cktweets"},
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
	// if err = bet.Valid(); err != nil {
	// 	return nil, err
	// } else {
	// 	err = db.UpsertBet(bet)
	// }
	err = db.UpsertBet(bet)
	return
}

// func CalcBetResult(bet *t.Bet) (err error) {
// 	fmt.Println("calc bet result ", bet.Id, bet.String())
// 	games := scraper.ScrapeCurrentGames()

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
