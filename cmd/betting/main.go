package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"bet-hound/cmd/db"
	"bet-hound/cmd/db/env"
	"bet-hound/cmd/nlp"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

const appConfigPath = "../db/env"
const appConfigName = "config"

var logger *log.Logger

const text = "I'll bet you that tevin coleman scores more ppr points than matt breida this week"

func main() {
	// Initialization
	logger = setUpLogger(env.LogPath(), "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading application config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	// Parse Text
	var sources []*t.Source
	var sourcePhrases []*t.Phrase
	nounPhrases, verbPhrases, _ := nlp.ParseText(text)

	if len(nounPhrases) < 2 {
		panic("Not enough noun phrases found!")
	}
	if len(verbPhrases) < 1 {
		panic("Not enough verb phrases found!")
	}

	// Find sources for noun phrases
	for _, nounPhrase := range nounPhrases {
		nounTxt := nounPhrase.AllText()
		foundSrcs, err := db.SearchSourceByName(strings.Join(nounTxt, " "), 1)
		if err != nil {
			fmt.Println(err)
		}
		if len(foundSrcs) > 0 {
			nounPhrase.Source = &foundSrcs[0]
			sourcePhrases = append(sourcePhrases, nounPhrase)
			sources = append(sources, &foundSrcs[0])
		}
	}
	if len(sourcePhrases) < 2 {
		panic("Not enough sources found!")
	}

	var metricPhrase, proposerSourcePhrase, recipientSourcePhrase *t.Phrase
	// Find Metric
	for _, n := range nounPhrases {
		nString := n.Word.Lemma
		isMetricStr := nString == "point" || nString == "pt" || nString == "yard" || nString == "yd" || nString == "touchdown" || nString == "td"
		if isMetricStr && n.Word.Children != nil && len(*n.Word.Children) > 1 {
			metricPhrase = n
			break
		}
	}
	if metricPhrase == nil {
		panic("Metric phrase not found!")
	}

	// Find Action
	var actionPhrase *t.Phrase
	for _, v := range verbPhrases {
		vString := v.Word.Lemma
		if vString == "score" || vString == "have" || vString == "gain" {
			for _, lemma := range v.AllLemmas() {
				if metricPhrase.Word.Lemma == lemma {
					actionPhrase = v
					break
				}
			}
		}
	}
	if actionPhrase == nil {
		panic("Action phrase not found!")
	}

	// Find Proposer Source
	for _, child := range *actionPhrase.Word.Children {
		for _, src := range sourcePhrases {
			if child.Text == src.Word.Text {
				proposerSourcePhrase = src
				break
			}
		}
	}
	if proposerSourcePhrase == nil {
		panic("Proposer source phrase not found!")
	}

	// Find Recipient Source
	for _, p := range nounPhrases {
		if p.Source != nil && p.Source != proposerSourcePhrase.Source {
			recipientSourcePhrase = p
			break
		}
	}
	// TODO : Calculate this through "breida" -> "than" -> "points"
	if proposerSourcePhrase == nil {
		panic("Recipient source phrase not found!")
	}

	fmt.Println("action word ", actionPhrase.AllLemmas())
	fmt.Println("metric word ", metricPhrase.AllLemmas())
	fmt.Println("proposer source ", proposerSourcePhrase.Word.Text)
	fmt.Println("recipient source ", recipientSourcePhrase.Word.Text)

	// Scrape Data
	// scraper.ScrapeSources()
	allGames := scraper.ScrapeThisWeeksGames()
	for _, game := range allGames {
		if *proposerSourcePhrase.Source.TeamFk == *game.HomeTeamFk {
			proposerSourcePhrase.HomeGame = game
		} else if *proposerSourcePhrase.Source.TeamFk == *game.AwayTeamFk {
			proposerSourcePhrase.AwayGame = game
		}

		if *recipientSourcePhrase.Source.TeamFk == *game.HomeTeamFk {
			recipientSourcePhrase.HomeGame = game
		} else if *recipientSourcePhrase.Source.TeamFk == *game.AwayTeamFk {
			recipientSourcePhrase.AwayGame = game
		}
	}
	if proposerSourcePhrase.Game() == nil {
		panic("Proposer source game not found!")
	}
	if recipientSourcePhrase.Game() == nil {
		panic("Recipient source game not found!")
	}

	fmt.Println("proposer source game", *proposerSourcePhrase.Game().Name)
	fmt.Println("recipient source game", *recipientSourcePhrase.Game().Name)

	// Build Bet
	// sources, err := db.SearchSourceByName("tevin colman", 1)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("found ", *sources[0].Name)
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
