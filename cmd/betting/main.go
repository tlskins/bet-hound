package main

import (
	"fmt"
	"log"
	"strings"
	// "net/http"
	// "encoding/json"
	// "io/ioutil"
	"os"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/nlp"
	t "bet-hound/cmd/types"
	// "bet-hound/cmd/scraper"
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

	// scraper.ScrapePlayers()

	tweet, err := db.FindTweet(1192258647562149888)
	if err != nil {
		fmt.Println("cant find tweet", err)
	}
	words := nlp.ParseText(*tweet.FullText)

	// Find players
	nouns := t.FindWords(&words, nil, &[]string{"NOUN"}, nil)
	players := []t.Player{}
	for _, n := range *nouns {
		children := t.FindWords(&words, &n.Index, &[]string{"NOUN"}, nil)
		if len(*children) > 0 {
			grouped := []t.Word{n}
			grouped = append(grouped, *children...)
			texts := t.WordsText(&grouped)
			t.ReverseStrings(texts)

			results := db.SearchPlayerByName(strings.Join(texts, " "), 1)
			if len(results) > 0 {
				fmt.Println("player:", results[0].Name)
				players = append(players, results[0])
			}
		}
	}

	// bet, err := nlp.ParseTweet(tweet)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("created bet fk", *bet.Fk)
	// fmt.Println("created bet", bet.Response())

	// bet, _ := db.FindBetByProposerCheckTweet("1192715899028922369")
	// bet, _ := db.FindBetById("c00716a6-4ad4-4f37-8708-db112c43fff2")
	// fmt.Println("text", bet.Text(), bet.BetStatus)

	// Reply to proposer check
	// logger.Println("reply to bet", *bet.Id, bet.Text())

	// replyTweetId := "1192702597905256448"
	// logger.Println("replyTweetId", replyTweetId)
	// bet, _ := db.FindBetByProposerCheckTweet(replyTweetId)
	// if err != nil {
	// 	logger.Println("err finding by proposer check tweet", err)
	// 	panic(err)
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
