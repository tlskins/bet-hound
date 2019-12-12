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
	// num_mod_txt1 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara scores ppr points this week?"
	num_mod_txt2 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara this week?"

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

	// Find Actions
	actions := t.SearchWords(&allWords, -1, -1, -1, []string{}, []string{"ACTION"})

	// Find Metrics
	// var metrics []*t.Word
	var lastDelim *t.Word
	for _, action := range actions {
		fmt.Printf("\nPROCESSING action word: %s %d\n", action.Text, action.Index)
		// lExpr := PlayerExpression{}
		// rExpr := PlayerExpression
		// op := t.OperatorPhrase{}
		var metric *t.Metric

		// Action child phrases
		fmt.Printf("Child phrases for action word: %s %d\n", action.Text, action.Index)
		words2D := t.SearchGroupedWords(&allWords, action.Index, -1, -1, false)

		// Each child phrase off action
		for _, words := range words2D {
			// Print child phrase
			for _, word := range words {
				fmt.Print(" " + word.Text)
			}
			fmt.Print("\n")

			// Find Deliminator
			if len(words) >= 1 {
				if words[0].BetComponent == "DELIMINATOR" {
					lastDelim = words[0]
					fmt.Printf("Found deliminator word: %s %d\n", words[0].Text, words[0].Index)
				}
			}

			// Find Metric
			if metric == nil {
				metricWord := t.SearchShallowestWord(&allWords, action.Index, -1, -1, []string{}, []string{"METRIC"})
				if metricWord != nil {
					metric = &t.Metric{Word: *metricWord}
					fmt.Printf("Found metric word: %s %d\n", metricWord.Text, metricWord.Index)
				}
			} else {
				// Find Metric Mods
				if words[len(words)-1].BetComponent == "METRIC" {
					for _, w := range words {
						if w.BetComponent == "METRIC_MOD" {
							fmt.Printf("Found metric mod: %s %d\n", w.Text, w.Index)
							metric.Modifiers = append(metric.Modifiers, *w)
						}
					}
				}
			}

			// Find Player
			notComponent := true
			playerNms := []*t.Word{}
			for _, word := range words {
				if word.PartOfSpeech.Tag == "NOUN" {
					playerNms = append(playerNms, word)
				}
				if len(word.BetComponent) > 0 {
					notComponent = false
				}
			}
			if notComponent && len(playerNms) > 0 {
				fmt.Print("Found player: ")
				for _, p := range playerNms {
					fmt.Print(" " + p.Text)
				}
				fmt.Print("\n")
			}
		}

		// Left or Right
		if lastDelim != nil && action.Index < lastDelim.Index {
			fmt.Println("Left expression")
		} else {
			fmt.Println("Right expression")
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
