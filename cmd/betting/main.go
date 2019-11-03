package main

import (
	"fmt"
	"log"
	"os"

	"bet-hound/cmd/db"
	"bet-hound/cmd/db/env"
	// "bet-hound/cmd/nlp"
	// t "bet-hound/cmd/types"
	// "bet-hound/cmd/scraper"
	m "bet-hound/pkg/mongo"
)

const appConfigPath = "../db/env"
const appConfigName = "config"

var logger *log.Logger

const text = "I'll bet you that tevin coleman scores more ppr points than matt breida this week"

func main() {
	logger = setUpLogger(env.LogPath(), "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading application config: %s \n", err)
	}
	defer env.Cleanup()

	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	// conn := env.MGOSession().Copy()
	// defer conn.Close()
	// c := conn.DB(env.MongoDb()).C(env.SourcesCollection())
	// m.CreateIndexKey(c, "name")

	// scraper.ScrapeSources()
	// games := scraper.ScrapeThisWeeksGames()
	// for _, game := range games {
	// 	fmt.Println(*game.Name)
	// }
	// groupedNouns, groupedVerbs := nlp.ParseText(text)
	sources, err := db.SearchSourceByName("tevin colman", 1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("found ", *sources[0].Name)
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
