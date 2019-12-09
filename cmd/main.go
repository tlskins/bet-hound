package main

import (
	"bet-hound/cmd/db"
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

const text = "yo fart face, do you wanna bet that Mike Evans scores more ppr points than Allen Robinson this week???"

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
	// loc, err := time.LoadLocation("America/Los_Angeles")
	// if err != nil {
	// 	fmt.Println(loc, err)
	// }
	// fmt.Println("game at: ", games[0].GameTime.In(loc))

	bets := db.FindPendingFinal()
	for _, bet := range *bets {
		fmt.Println(bet.Equation.Text(), b.CalcBetResult(bet))
	}

	// fmt.Println(bet.Equation.LeftExpression.Game.GameTime.In(loc))
	// fmt.Println(bet.FinalizedAt().In(loc))

	// bet, err := db.FindBetById("9b6247a8-6653-4e86-845e-3b8c25296331")
	// // result := b.CalcBetResult(bet)
	// // fmt.Println(result)

	// game := bet.Equation.LeftExpression.Game
	// fmt.Println("game time ", game.GameTime.String())
	// gameEnd := game.GameTime.Add(time.Hour * 6)
	// fmt.Println("game end ", gameEnd.String())
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
