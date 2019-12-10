package main

import (
	// "bet-hound/cmd/db"
	"fmt"
	"log"
	"os"
	// "time"

	b "bet-hound/cmd/betting"
	"bet-hound/cmd/env"
	// "bet-hound/cmd/scraper"
	// t "bet-hound/cmd/types"
	// "bet-hound/cmd/twitter"
	m "bet-hound/pkg/mongo"
)

const appConfigPath = "./env"
const appConfigName = "config"

var logger *log.Logger

const text = "@bettybetbot @richayelfuego yo richardo u wanna bet that Alshon Jeffery scores more ppr points that Saquon Barkley this week?"

func main() {
	// Initialization
	logger = setUpLogger(env.LogPath(), "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading application config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	// games := scraper.ScrapeThisWeeksGames()
	// // bet, err := db.FindBetById("5a12fcb9-aff3-4f16-b8e8-c8b34e4a0942")
	// loc, err := time.LoadLocation("America/New_York")
	// if err != nil {
	// 	fmt.Println(loc, err)
	// }
	// fmt.Println(time.Date(2019, 12, 9, 12, 0, 0, 0, loc).UTC())

	// fmt.Println("game at: ", games[0].GameTime.In(loc))

	// tweet, _ := db.FindTweet("501399114")
	// fmt.Println(*tweet)
	_, eq := b.BuildEquationFromText(text)
	fmt.Println(eq)
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
