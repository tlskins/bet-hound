package main

import (
	"fmt"
	"log"
	// "strings"
	// "net/http"
	// "encoding/json"
	// "io/ioutil"
	"os"

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

	opPhrase, leftMetric := nlp.FindOperatorPhrase(&words)

	leftPlayerExpr := nlp.FindLeftPlayerExpr(&words, opPhrase, leftMetric)

	fmt.Println("left player expr: ", leftPlayerExpr.Player.Name, leftPlayerExpr.Metric.Word.Text)

	rightPlayerExpr := nlp.FindRightPlayerExpr(&words, opPhrase, leftMetric)

	fmt.Println("right player expr: ", rightPlayerExpr.Player.Name)

	eq := t.Equation{
		LeftExpression:  *leftPlayerExpr,
		RightExpression: *rightPlayerExpr,
		Operator:        *opPhrase,
	}
	fmt.Println(
		"equation: ",
		eq.LeftExpression.Player.Name,
		eq.Operator.ActionWord.Text,
		eq.Operator.OperatorWord.Text,
		leftPlayerExpr.Metric.Word.Text,
		leftPlayerExpr.Metric.Modifiers,
		leftPlayerExpr.EventTime.Word.Text,
		leftPlayerExpr.EventTime.Modifiers,
		rightPlayerExpr.Player.Name,
	)

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
