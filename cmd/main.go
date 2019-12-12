package main

import (
	// "bet-hound/cmd/db"
	"fmt"
	"log"
	"os"
	"strings"
	// "time"

	// b "bet-hound/cmd/betting"
	"bet-hound/cmd/env"
	"bet-hound/cmd/nlp"
	// "bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
	// "bet-hound/cmd/twitter"
	m "bet-hound/pkg/mongo"
)

const appConfigPath = "./env"
const appConfigName = "config"

var logger *log.Logger

func main() {
	// Initialization
	logger = setUpLogger(env.LogPath(), "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading application config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	// Text samples
	// pt_to_vrb_txt := "@bettybetbot @richayelfuego yo richardo u wanna bet that Alshon Jeffery scores more ppr points that Saquon Barkley this week?"
	// name_matching_txt := "@bettybetbot @richayelfuego bet you that juju scores more ppr points than AJ Brown this week?"
	// num_mod_txt1 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara this week?"
	num_mod_txt2 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara scores ppr points this week?"

	// tweet, nil := db.FindTweet("1204576588387373056")
	// tweet.FullText = &num_mod_txt
	// err, bet := b.BuildBetFromTweet(tweet)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// lMetric, rMetric := bet.Equation.MetricString()
	// fmt.Println("bet", bet.Equation.Operator, lMetric, rMetric)

	fmt.Println("start")
	txt := strings.TrimSpace(nlp.RemoveReservedTwitterWords(num_mod_txt2))
	allWords := nlp.ParseText(txt)

	// Expression Indices
	delims := t.SearchWords(&allWords, -1, -1, -1, []string{}, []string{"DELIMINATOR"})
	exprIdxs := [][]int{}
	for i, delim := range delims {
		// Find index of next delim
		var nxtDelimIdx int
		if i+1 < len(delims) {
			nxtDelimIdx = delims[i+1].Index
		} else {
			nxtDelimIdx = -1
		}

		if i == 0 {
			exprIdxs = append(exprIdxs, []int{-1, delim.Index})
			exprIdxs = append(exprIdxs, []int{delim.Index, nxtDelimIdx})
		} else {
			exprIdxs = append(exprIdxs, []int{delims[i-1].Index, delim.Index})
			exprIdxs = append(exprIdxs, []int{delim.Index, nxtDelimIdx})
		}
	}

	// Build Expressions
	for _, idxs := range exprIdxs {
		stIdx := idxs[0]
		endIdx := idxs[1]
		root := allWords[0]
		if stIdx != -1 {
			root = allWords[stIdx]
		}
		fmt.Printf("\nBuilding expression root '%s' start %d to end %d\n", root.Text, stIdx, endIdx)

		var action, operator *t.Word
		var metric *t.Metric

		// Child Phrases
		wordsPaths := t.SearchGroupedWords(&allWords, root.Index, stIdx, endIdx, false)
		for _, words := range wordsPaths {
			// Print child phrase
			for _, word := range words {
				fmt.Print(" " + word.Text)
			}
			fmt.Print("\n")

			// Find action
			if action == nil {
				actionWord := t.SearchFirstWord(&allWords, stIdx, endIdx, []string{}, []string{"ACTION"})
				if actionWord != nil {
					action = actionWord
					fmt.Printf("Found action word: %s %d\n", action.Text, action.Index)

					// Find Metric
					if metric == nil {
						metricWord := t.SearchShallowestWord(&allWords, action.Index, -1, -1, []string{}, []string{"METRIC"})
						if metricWord != nil {
							metric = &t.Metric{Word: *metricWord}
							fmt.Printf("Found metric word: %s %d\n", metric.Word.Text, metric.Word.Index)
							metricPaths := t.SearchGroupedWords(&allWords, metricWord.Index, stIdx, endIdx, true)
							for _, words := range metricPaths {
								// Find Metric Mods
								if words[len(words)-1].BetComponent == "METRIC" {
									for _, m := range words {
										if m.BetComponent == "METRIC_MOD" {
											fmt.Printf("Found metric mod: %s %d\n", m.Text, m.Index)
											metric.Modifiers = append(metric.Modifiers, *m)
										}
									}
								}

								// Find Operator
								if operator == nil {
									opWord := t.SearchShallowestWord(&allWords, metricWord.Index, stIdx, endIdx, []string{}, []string{"OPERATOR"})
									if opWord != nil {
										operator = opWord
										fmt.Printf("Found operator word: %s %d\n", operator.Text, operator.Index)
									}
								}
							}
						}
					}
					// 	// Find Metric Mods
					// 	if words[len(words)-1].BetComponent == "METRIC" {
					// 		for _, w := range words {
					// 			if w.BetComponent == "METRIC_MOD" {
					// 				fmt.Printf("Found metric mod: %s %d\n", w.Text, w.Index)
					// 				metric.Modifiers = append(metric.Modifiers, *w)
					// 			}
					// 		}
					// 	}
					// }
				}
			}

			// Find Player
			playerWds := []*t.Word{}
			for _, word := range words {
				if word.PartOfSpeech.Tag == "NOUN" && len(word.BetComponent) == 0 {
					playerWds = append(playerWds, word)
				}
			}
			if len(playerWds) > 0 {
				fmt.Print("Found player: ")
				for _, w := range playerWds {
					fmt.Print(" " + w.Text)
				}
				fmt.Print("\n")
			}
		}
	}
}

func setUpLogger(logPath, defaultPath string) *log.Logger {
	if logPath == "" {
		logPath = defaultPath
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(f, "", 0)
}
