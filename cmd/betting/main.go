package main

import (
	"fmt"
	"log"
	// "net/http"
	// "encoding/json"
	// "io/ioutil"
	"os"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/nlp"
	// "bet-hound/cmd/twitter"
	m "bet-hound/pkg/mongo"
)

const appConfigPath = "../env"
const appConfigName = "config"

var logger *log.Logger

const text = "yo fart face, do wanna bet that Emmanuel Sanders scores more points than Allen Robinson this week???"

func main() {
	// Initialization
	logger = setUpLogger(env.LogPath(), "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading application config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	// // get twitter
	// url := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?tweet_mode=extended&id=%s", "1191855424342908928")
	// client := twitter.CreateClient()
	// resp, err := client.Get(url)
	// if err != nil {
	// 	fmt.Println("err", err)
	// }
	// defer resp.Body.Close()

	// body, _ := ioutil.ReadAll(resp.Body)
	// var data map[string]interface{}
	// if err := json.Unmarshal([]byte(body), &data); err != nil {
	// 	fmt.Println("err", err)
	// 	panic(err)
	// }
	// fmt.Println("data", data)

	// scraper.ScrapeSources()
	tweet, err := db.FindTweet(1192258647562149888)
	if err != nil {
		fmt.Println("cant find tweet", err)
	}
	bet, err := nlp.ParseTweet(tweet)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("created bet fk", *bet.Fk)
	fmt.Println("created bet", bet.Response())
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
