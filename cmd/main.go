package main

import (
	"fmt"
	"log"
	"os"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"

	"bet-hound/pkg/helpers"
	m "bet-hound/pkg/mongo"
)

const appConfigPath = "./env"
const appConfigName = "config"

var logger *log.Logger

func main() {
	// Initialization
	logger = setUpLogger(env.LogPath(), "logs.log")
	if err := env.Init(appConfigName, appConfigPath); err != nil {
		panic(err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	boards, err := db.CurrentLeaderBoards()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(helpers.PrettyPrint(boards))
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
