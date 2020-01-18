package main

import (
	// "fmt"
	"log"
	"os"
	// "strings"
	// "time"

	// b "bet-hound/cmd/betting"
	// "bet-hound/cmd/cron"
	"bet-hound/cmd/env"
	// "bet-hound/cmd/nlp"
	"bet-hound/cmd/scraper"
	// t "bet-hound/cmd/types"
	// "bet-hound/cmd/db"
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

	scraper.ScrapeGames(2019, 19)

	// lgSettings := cron.InitLeagueSettings()
	// fmt.Println("settings: ", lgSettings)

	// game, _ := db.FindGameById("201919SFOMIN")
	// gameLog, err := scraper.ScrapeGameLog(game.Url)
	// fmt.Println(err, gameLog)

	// scraper.ScrapeGames(2019, 19)

	// games, err := db.GetResultReadyGames()
	// for _, g := range games {
	// 	fmt.Println(g.Name)
	// }

	// if results, err := cron.CheckGameResults(lgSettings); err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	for _, game := range *results {
	// 		bets, err := db.FindAcceptedBetsByGame(game.Id)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			continue
	// 		}
	// 		for _, bet := range *bets {
	// 			fmt.Println(bet.String())
	// 			b.EvaluateBet(bet, game)
	// 		}
	// 	}
	// }

	// bet, err := db.FindBetById("26a9c841-1846-4077-8000-33e179c7eb71")
	// fmt.Println(err)

	// Text samples
	// pt_to_vrb_txt := "@bettybetbot @richayelfuego yo richardo u wanna bet that Alshon Jeffery scores more ppr points that Saquon Barkley this week?"
	// name_matching_txt := "@bettybetbot @richayelfuego bet you that juju scores more ppr points than AJ Brown this week?"
	// num_mod_txt1 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara this week?"
	// num_mod_txt2 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more ppr points than Alvin Kamara scores ppr points this week?"
	// num_mod_txt3 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery scores 5.6 more points than Alvin Kamara in ppr this week?"
	// num_mod_txt4 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery and Adrian Peterson score 5.6 more ppr points than Alvin Kamara this week?"
	// num_mod_txt5 := "@bettybetbot @richayelfuego bet you that Alshon Jeffery and Carson Wentz score 5.6 more ppr points than Alvin Kamara, James Washington, and Christian Kirk?"

	// tweet, nil := db.FindTweet("1206273109411524609")
	// // tweet.FullText = &num_mod_txt5
	// err, bet := b.BuildBetFromTweet(tweet)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// bet.Id = "test"
	// bet.PostProcess()
	// valid := bet.Valid()

	// fmt.Println("bet ", bet.Description(), bet.ExpiresAt.String(), bet.FinalizedAt.String())
	// bet.AcceptBy(bet.Proposer.IdStr, "proposer_reply_fk")
	// fmt.Println("bet ", bet.Description(), bet.ProposerReplyFk)
	// db.UpsertBet(bet)

	// bet, _ := db.FindBetById("5933b58a-5f7c-4fbd-a8e9-15d114d4cb56")
	// b.CalcBetResult(bet)
	// fmt.Println(bet.BetResult.Response)

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
