package main

import (
	"bet-hound/cmd/db"
	"fmt"
	"log"
	"os"
	// "strings"
	// "time"

	b "bet-hound/cmd/betting"
	"bet-hound/cmd/env"
	// "bet-hound/cmd/nlp"
	// "bet-hound/cmd/scraper"
	// t "bet-hound/cmd/types"
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
	// num_mod_txt2 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara scores ppr points this week?"
	// num_mod_txt3 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more points than Alvin Kamara in ppr this week?"
	// num_mod_txt4 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery and Adrian Peterson score 5.6 more ppr points than Alvin Kamara this week?"
	num_mod_txt5 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery and Carson Wentz score 5.6 more ppr points than Alvin Kamara, James Washington, and Christian Kirk?"

	tweet, nil := db.FindTweet("1204576588387373056")
	tweet.FullText = &num_mod_txt5
	err, bet := b.BuildBetFromTweet(tweet)
	if err != nil {
		fmt.Println(err)
	}
	bet.Id = "test"
	bet.PostProcess()
	fmt.Println("bet ", bet.Description(), bet.ExpiresAt.String(), bet.FinalizedAt.String())
	bet.AcceptBy(bet.Proposer.IdStr, "proposer_reply_fk")
	fmt.Println("bet ", bet.Description(), bet.ProposerReplyFk)
	db.UpsertBet(bet)

	// eqs, err := b.BuildEquationsFromText(num_mod_txt5)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// for _, q := range eqs {
	// 	fmt.Printf("%s\n", q.Description())
	// }
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
