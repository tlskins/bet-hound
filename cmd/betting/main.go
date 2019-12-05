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

	// tweet, err := db.FindTweet(1192258647562149888)
	// if err != nil {
	// 	fmt.Println("cant find tweet", err)
	// }
	words := nlp.ParseText(text)

	// Find players
	nouns := t.FindWords(&words, nil, &[]string{"NOUN"}, nil)
	playerExprs := []t.PlayerExpression{}
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
				expr := t.PlayerExpression{
					Player: results[0],
				}
				playerExprs = append(playerExprs, expr)
			}
		}
	}
	if len(playerExprs) < 2 {
		panic("Not enough players found!")
		// return bet, fmt.Errorf("Not enough sources found!")
	}

	actionWords := nlp.FindActions(words)
	fmt.Println("action words: ", t.WordsText(&actionWords))

	operatorPhrases := nlp.FindOperatorPhrases(words, actionWords)
	for _, p := range operatorPhrases {
		fmt.Println("operator phrase: ", p.ActionWord.Lemma, p.OperatorWord.Lemma)
	}

	for _, op := range operatorPhrases {
		for _, player := range playerExprs {
			FindWords(&words, &player.Word.Index, nil, nil)
		}
	}
	// fmt.Println("operator word: ", operatorWord.Text)
	// fmt.Println("metric: ", metric.Text, metric.Modifiers)

	// Find Metric
	// var metric *t.Metric
	// for _, n := range *nouns {
	// 	str := n.Lemma
	// 	isMetricStr := str == "point" || str == "pt" || str == "yard" || str == "yd" || str == "touchdown" || str == "td"
	// 	if !isMetricStr {
	// 		continue
	// 	}
	// 	children := t.FindWords(&words, &n.Index, &[]string{"NOUN", "ADJ"}, nil)
	// 	if len(*children) > 0 {
	// 		metric = &t.Metric{
	// 			Text:      n.Text,
	// 			Lemma:     n.Lemma,
	// 			Modifiers: t.WordsLemmas(children),
	// 		}
	// 	}
	// }
	// if metric == nil {
	// 	panic("betting metric not found!")
	// 	// return bet, fmt.Errorf("Metric phrase not found!")
	// } else {
	// 	fmt.Println("Metric: ", metric.Lemma, metric.Modifiers)
	// 	for _, p := range playerExprs {
	// 		p.Metric = metric
	// 	}
	// }

	// Find action word
	// var actionWord, operatorWord *string
	// var metric *t.Metric

	// var opPhrase *t.OperatorPhrase
	// verbs := t.FindWords(&words, nil, &[]string{"VERB"}, nil)
	// for _, v := range *verbs {
	// 	str := v.Lemma
	// 	if str == "score" || str == "have" || str == "gain" {
	// 		vChildren := t.FindWords(&words, &v.Index, &[]string{"NOUN"}, nil)
	// 		if len(*vChildren) == 0 {
	// 			continue
	// 		}
	// 		// found action
	// 		fmt.Println("action: ", v.Lemma, t.WordsLemmas(nChildren))
	// 		actionWord = &v
	// 		for _, n := range *vChildren {
	// 			str = n.Lemma
	// 			if str == "point" || str == "pt" || str == "yard" || str == "yd" || str == "touchdown" || str == "td" {
	// 				aChildren := t.FindWords(&words, &n.Index, &[]string{"NOUN", "ADJ"}, nil)
	// 				if len(*aChildren) > 0 {
	// 					// found metric
	// 					metric = &t.Metric{
	// 						Text:      n.Text,
	// 						Lemma:     n.Lemma,
	// 						Modifiers: []string{},
	// 					}
	// 					for _, a := range aChildren {
	// 						str = m.Lemma
	// 						if str == "more" || str == "great" || str == "few" || str == "less" {
	// 							// found operator
	// 							operatorWord = &m
	// 						} else if m.PartOfSpeech.Tag == "NOUN" || m.PartOfSpeech.Tag == "ADJ" {
	// 							// found metric modifier
	// 							metric.Modifiers = append(metric.Modifiers, m.Text)
	// 						}
	// 					}
	// 				}
	// 			}
	// 			if metric != nil && actionWord != nil && operatorWord != nil {
	// 				break
	// 			}
	// 		}
	// 	}
	// }

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
