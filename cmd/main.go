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

// Points comes up as verb
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

	err, eq := b.BuildEquationFromText(text)
	if err != nil {
		fmt.Println(err)
	}
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
