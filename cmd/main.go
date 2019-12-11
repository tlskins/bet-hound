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
	num_mod_txt1 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara scores ppr points this week?"
	// num_mod_txt2 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara this week?"

	// tweet, nil := db.FindTweet("1204576588387373056")
	// tweet.FullText = &num_mod_txt
	// err, bet := b.BuildBetFromTweet(tweet)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// lMetric, rMetric := bet.Equation.MetricString()
	// fmt.Println("bet", bet.Equation.Operator, lMetric, rMetric)

	fmt.Println("start")
	txt := strings.TrimSpace(nlp.RemoveReservedTwitterWords(num_mod_txt1))
	words := nlp.ParseText(txt)
	actions := t.SearchWords(&words, -1, -1, -1, []string{}, []string{"ACTION"})
	var actionWords [][]*t.Word
	for _, action := range actions {
		recWords := t.SearchGroupedWords(&words, action.Index, -1, -1)
		for _, r := range recWords {
			actionWords = append(actionWords, r)
		}
	}
	for i, a := range actionWords {
		fmt.Print("Actions word ", i)
		for _, n := range a {
			fmt.Print(" " + n.Text)
		}
		fmt.Print("\n")
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
