package betting

import (
	"bet-hound/cmd/db"
	"bet-hound/cmd/nlp"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
	"fmt"
	"github.com/satori/go.uuid"
	"math"
	"strings"
	"time"
)

func BuildBetFromTweet(tweet *t.Tweet) (err error, bet *t.Bet) {
	text := strings.TrimSpace(nlp.RemoveReservedTwitterWords(tweet.GetText()))
	eqs, err := BuildEquationsFromText(text)
	if err != nil {
		return err, nil
	}
	if len(tweet.Recipients()) == 0 {
		return fmt.Errorf("Not enough recipients!"), nil
	}
	recipient := tweet.Recipients()[0]

	bet = &t.Bet{
		Id:        uuid.NewV4().String(),
		SourceFk:  tweet.IdStr,
		Proposer:  tweet.User,
		Recipient: recipient,
		BetStatus: t.BetStatusFromString("Pending Proposer"),
		Equations: eqs,
	}
	bet.PostProcess()
	valid := bet.Valid()
	return valid, bet
}

func BuildEquationsFromText(text string) (eqs []*t.Equation, err error) {
	allWords := nlp.ParseText(text)
	playerWords := nlp.FindPlayerWords(&allWords)
	currentGames := scraper.ScrapeThisWeeksGames()

	// Build Equations
	actionEqsMap := make(map[int]*t.Equation)
	for _, pw := range playerWords {
		// Find Player
		playerWord := pw[len(pw)-1]
		lemmas := nlp.WordsLemmas(&pw)
		player := db.SearchPlayerByName(strings.Join(lemmas, " "))
		if player == nil {
			fmt.Printf("Player not found.\n")
			continue
		}
		// Find game
		game := findGameByFk(&currentGames, player.TeamFk)
		if game == nil {
			fmt.Printf("Game not found for %s.\n", player.Name)
			continue
		}
		// Find action
		action := nlp.SearchLastParent(&allWords, playerWord.Index, -1, -1, []string{}, []string{"ACTION"})
		if action == nil {
			fmt.Printf("No action found for %s.\n", player.Name)
			continue
		}

		var eq *t.Equation
		var delimiter *t.Word
		// Get / Build equation
		if actionEqsMap[action.Index] != nil {
			eq = actionEqsMap[action.Index]
			delimiter = eq.Delimiter
		} else {
			delimiter = nlp.SearchFirstChild(&allWords, action.Index, -1, -1, []string{}, []string{"DELIMITER"})
			if delimiter == nil {
				fmt.Printf("No delimiter found for %s.\n", player.Name)
				continue
			}

			operator := nlp.SearchFirstChild(&allWords, action.Index, playerWord.Index, -1, []string{}, []string{"OPERATOR"})
			if operator == nil {
				fmt.Printf("No operator found for %s.\n", player.Name)
				continue
			}

			metricWord := nlp.SearchFirstChild(&allWords, action.Index, -1, -1, []string{}, []string{"METRIC"})
			var metric *t.Metric
			if metricWord == nil {
				fmt.Printf("No metric found for %s.\n", player.Name)
				continue
			} else {
				mods := nlp.SearchChildren(&allWords, metricWord.Index, -1, -1, []string{}, []string{"METRIC_MOD"})
				metric = &t.Metric{
					Word:      metricWord,
					Modifiers: mods,
				}
			}

			eq = &t.Equation{
				Action:    action,
				Metric:    metric,
				Delimiter: delimiter,
				Operator:  operator,
			}
			actionEqsMap[action.Index] = eq
		}

		// Build Expression
		expr := t.PlayerExpression{
			Player: player,
			Game:   game,
		}

		if playerWord.Index < delimiter.Index {
			eq.LeftExpressions = append(eq.LeftExpressions, &expr)
		} else {
			eq.RightExpressions = append(eq.RightExpressions, &expr)
		}
	}

	for _, eq := range actionEqsMap {
		eqs = append(eqs, eq)
	}
	if len(eqs) == 0 {
		return eqs, fmt.Errorf("No equations found!.")
	}
	return eqs, nil
}

func CalcBetResult(bet *t.Bet) (err error) {
	fmt.Println("calc bet result ", bet.Id, bet.Description())
	games := scraper.ScrapeThisWeeksGames()

	responses := []string{}
	proposerWins := true
	for _, eq := range bet.Equations {
		eqResult, err := calcEquationResult(eq, &games)
		if err != nil {
			return err
		}
		responses = append(responses, fmt.Sprintf("%s (%t)", eq.ResultDescription(), *eqResult))
		proposerWins = proposerWins && *eqResult
	}

	var wUsr, lUsr t.User
	if proposerWins {
		wUsr = bet.Proposer
		lUsr = bet.Recipient
	} else {
		wUsr = bet.Recipient
		lUsr = bet.Proposer
	}
	bet.BetResult = &t.BetResult{
		Winner: wUsr,
		Loser:  lUsr,
		Response: fmt.Sprintf("Congrats @%s you beat @%s! '%s'",
			wUsr.ScreenName,
			lUsr.ScreenName,
			strings.Join(responses, ", "),
		),
		DecidedAt: time.Now(),
	}
	bet.BetStatus = t.BetStatusFromString("Final")
	return nil
}

// helpers

func findGameByFk(games *[]*t.Game, teamFk string) *t.Game {
	for _, g := range *games {
		if g.AwayTeamFk == teamFk || g.HomeTeamFk == teamFk {
			return g
		}
	}
	return nil
}

func calcEquationResult(eq *t.Equation, games *[]*t.Game) (*bool, error) {
	// Process each expression
	allExprs := [][]*t.PlayerExpression{eq.LeftExpressions, eq.RightExpressions}
	errs := []string{}
	for _, exprs := range allExprs {
		for _, expr := range exprs {
			err := calcExpressionResult(expr, games, eq.Metric)
			if err != nil {
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf(strings.Join(errs, ""))
	}

	// Record result
	lTtl := calcExpressionsTotal(&eq.LeftExpressions)
	rTtl := calcExpressionsTotal(&eq.RightExpressions)
	fixedMod := eq.Metric.FixedValueMod()
	var result bool
	if eq.Operator.Lemma == "more" {
		// Add fixed mods to equation
		if fixedMod != nil {
			*lTtl += *fixedMod
		}
		result = *lTtl > *rTtl
	} else if eq.Operator.Lemma == "less" || eq.Operator.Lemma == "few" {
		if fixedMod != nil {
			*lTtl -= *fixedMod
		}
		result = *lTtl < *rTtl
	}
	eq.Result = &result
	return &result, nil
}

func calcExpressionsTotal(expressions *[]*t.PlayerExpression) *float64 {
	total := 0.0
	for _, e := range *expressions {
		if e.Value == nil {
			return nil
		}
		total += *e.Value
	}

	return &total
}

func calcExpressionResult(expr *t.PlayerExpression, games *[]*t.Game, metric *t.Metric) (err error) {
	gm := findGameByFk(games, expr.Player.TeamFk)
	log := scraper.ScrapeGameLog(gm)
	value := calcPlayerGameValue(&log, expr.Player, metric)

	if value == nil {
		return fmt.Errorf("Unable to determine score for %s.", expr.Description())
	} else {
		expr.Value = value
		return nil
	}
}

func calcPlayerGameValue(log *map[string]*t.GameStat, player *t.Player, metric *t.Metric) *float64 {
	l := *log
	score := 0.0
	if l[player.Fk] == nil {
		return &score
	}
	score += float64(l[player.Fk].PassYd) * 0.04
	score += float64(l[player.Fk].PassTd) * 4.0
	score -= float64(l[player.Fk].PassInt) * 2.0
	// score -= float64(l[player.Fk].PassSackedYd) / 10.0
	score += float64(l[player.Fk].RushYd) * 0.1
	score += float64(l[player.Fk].RushTd) * 6.0
	score += float64(l[player.Fk].Rec) * metric.PPR()
	score += float64(l[player.Fk].RecYd) * 0.1
	score += float64(l[player.Fk].RecTd) * 6.0
	score -= float64(l[player.Fk].FumbleLost) * 2.0
	score = math.Ceil(score*10) / 10
	return &score
}
