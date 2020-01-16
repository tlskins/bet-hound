package betting

import (
	"fmt"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
)

func TweetBetProposal(bet *t.Bet) (*t.Tweet, error) {
	client := env.TwitterClient()
	txt := fmt.Sprintf("@%s %s has proposed a bet that: %s. Do you accept?", bet.Recipient.TwitterUser.ScreenName, bet.Proposer.Name, bet.String())
	resp, err := client.SendTweet(txt, nil)
	if err != nil {
		return nil, err
	}
	bet.AcceptFk = resp.IdStr
	if err = db.UpsertTweet(resp); err != nil {
		fmt.Println(err)
	}
	if err = db.UpsertBet(bet); err != nil {
		return nil, err
	}
	return resp, nil
}

// func BuildBetFromTweet(tweet *t.Tweet) (err error, bet *t.Bet) {
// 	text := strings.TrimSpace(nlp.RemoveReservedTwitterWords(tweet.GetText()))
// 	eqs, err := BuildEquationsFromText(text)
// 	if err != nil {
// 		return err, nil
// 	}
// 	if len(tweet.Recipients()) == 0 {
// 		return fmt.Errorf("Not enough recipients!"), nil
// 	}
// 	recipient := tweet.Recipients()[0]

// 	bet = &t.Bet{
// 		Id:        uuid.NewV4().String(),
// 		SourceFk:  tweet.IdStr,
// 		Proposer:  tweet.User,
// 		Recipient: recipient,
// 		BetStatus: t.BetStatusFromString("Pending Proposer"),
// 		Equations: eqs,
// 	}
// 	bet.PostProcess()
// 	valid := bet.Valid()
// 	return valid, bet
// }

// func BuildEquationsFromText(text string) (eqs []*t.Equation, err error) {
// 	allWords := nlp.ParseText(text)
// 	playerWords := nlp.FindPlayerWords(&allWords)
// 	currentGames := scraper.ScrapeThisWeeksGames()

// 	// Build Equations
// 	actionEqsMap := make(map[int]*t.Equation)
// 	for _, pw := range playerWords {
// 		// Find Player
// 		playerWord := pw[len(pw)-1]
// 		lemmas := nlp.WordsLemmas(&pw)
// 		player := db.SearchPlayerByName(strings.Join(lemmas, " "))
// 		if player == nil {
// 			fmt.Printf("Player not found.\n")
// 			continue
// 		}
// 		// Find game
// 		game := findGameByFk(&currentGames, player.TeamFk)
// 		if game == nil {
// 			fmt.Printf("Game not found for %s.\n", player.Name)
// 			continue
// 		}
// 		// Find action
// 		action := nlp.SearchLastParent(&allWords, playerWord.Index, -1, -1, []string{}, []string{"ACTION"})
// 		if action == nil {
// 			fmt.Printf("No action found for %s.\n", player.Name)
// 			continue
// 		}

// 		var eq *t.Equation
// 		var delimiter *t.Word
// 		// Get / Build equation
// 		if actionEqsMap[action.Index] != nil {
// 			eq = actionEqsMap[action.Index]
// 			delimiter = eq.Delimiter
// 		} else {
// 			delimiter = nlp.SearchFirstChild(&allWords, action.Index, -1, -1, []string{}, []string{"DELIMITER"})
// 			if delimiter == nil {
// 				fmt.Printf("No delimiter found for %s.\n", player.Name)
// 				continue
// 			}

// 			operator := nlp.SearchFirstChild(&allWords, action.Index, playerWord.Index, -1, []string{}, []string{"OPERATOR"})
// 			if operator == nil {
// 				fmt.Printf("No operator found for %s.\n", player.Name)
// 				continue
// 			}

// 			metricWord := nlp.SearchFirstChild(&allWords, action.Index, -1, -1, []string{}, []string{"METRIC"})
// 			var metric *t.Metric
// 			if metricWord == nil {
// 				fmt.Printf("No metric found for %s.\n", player.Name)
// 				continue
// 			} else {
// 				mods := nlp.SearchChildren(&allWords, metricWord.Index, -1, -1, []string{}, []string{"METRIC_MOD"})
// 				metric = &t.Metric{
// 					Word:      metricWord,
// 					Modifiers: mods,
// 				}
// 			}

// 			eq = &t.Equation{
// 				Action:    action,
// 				Metric:    metric,
// 				Delimiter: delimiter,
// 				Operator:  operator,
// 			}
// 			actionEqsMap[action.Index] = eq
// 		}

// 		// Build Expression
// 		expr := t.PlayerExpression{
// 			Player: player,
// 			Game:   game,
// 		}

// 		if playerWord.Index < delimiter.Index {
// 			eq.LeftExpressions = append(eq.LeftExpressions, &expr)
// 		} else {
// 			eq.RightExpressions = append(eq.RightExpressions, &expr)
// 		}
// 	}

// 	for _, eq := range actionEqsMap {
// 		eqs = append(eqs, eq)
// 	}
// 	if len(eqs) == 0 {
// 		return eqs, fmt.Errorf("No equations found!.")
// 	}
// 	return eqs, nil
// }
