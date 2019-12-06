package main

import (
	"fmt"
	"log"
	// "strings"
	// "net/http"
	// "encoding/json"
	// "io/ioutil"
	"os"

	// "bet-hound/cmd/db"
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

const text = "yo fart face, do wanna bet that Mike Evans scores more ppr points than Allen Robinson this week???"

func main() {
	// Initialization
	logger = setUpLogger(env.LogPath(), "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading application config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	words := nlp.ParseText(text)

	actionWords := nlp.FindActions(&words)
	fmt.Println("action words: ", t.WordsText(&actionWords))

	opPhrase := &t.OperatorPhrase{}
	leftMetricWord := &t.Word{}
	for _, action := range actionWords {
		opPhrase, leftMetricWord = nlp.FindOperatorPhrase(&words, action)
		if opPhrase != nil {
			fmt.Println("operator phrase: ", opPhrase.ActionWord.Lemma, opPhrase.OperatorWord.Lemma)
			break
		}
	}
	if opPhrase == nil || leftMetricWord == nil {
		fmt.Println("no op phrase or left metric!")
		return
	}
	fmt.Println("left metric word: ", leftMetricWord.Text)

	nouns := t.FindWords(&words, opPhrase.ActionWord.Index, []string{"NOUN"}, []string{leftMetricWord.Text})
	for _, noun := range nouns {
		fmt.Println("noun: ", noun.Text, leftMetricWord)
	}

	// Find players
	// nouns := t.FindWords(&words, nil, &[]string{"NOUN"}, nil)
	// playerExprs := []t.PlayerExpression{}
	// for _, n := range *nouns {
	// 	children := t.FindWords(&words, &n.Index, &[]string{"NOUN"}, nil)
	// 	if len(*children) > 0 {
	// 		grouped := []t.Word{n}
	// 		grouped = append(grouped, *children...)
	// 		texts := t.WordsText(&grouped)
	// 		t.ReverseStrings(texts)
	// 		results := db.SearchPlayerByName(strings.Join(texts, " "), 1)
	// 		if len(results) > 0 {
	// 			fmt.Println("player:", results[0].Name)
	// 			expr := t.PlayerExpression{
	// 				Player: results[0],
	// 			}
	// 			playerExprs = append(playerExprs, expr)
	// 		}
	// 	}
	// }
	// if len(playerExprs) < 2 {
	// 	panic("Not enough players found!")
	// 	// return bet, fmt.Errorf("Not enough sources found!")
	// }

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
