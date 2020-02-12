package main

import (
	// "fmt"
	"fmt"
	"log"
	"os"
	"time"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"

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

	now := time.Now().In(env.TimeZone())
	weekEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, env.TimeZone())
	if wd := weekEnd.Weekday(); wd == time.Sunday {
		weekEnd = weekEnd.AddDate(0, 0, -6)
	} else {
		weekEnd = weekEnd.AddDate(0, 0, -int(wd)+1)
	}
	fmt.Println(weekEnd.String())
	weekStart := weekEnd.AddDate(0, 0, -7)
	fmt.Println(weekStart.String())

	db.BuildLeaderBoard(&weekStart, &weekEnd, "nba")
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
