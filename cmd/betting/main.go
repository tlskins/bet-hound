package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"bet-hound/cmd/db"
	"bet-hound/cmd/db/env"
	"bet-hound/cmd/nlp"
	t "bet-hound/cmd/types"
	// "bet-hound/cmd/scraper"
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

	var proposerSourcePhrase, recipientSourcePhrase *t.Phrase
	for _, verb := range verbPhrases {
		if (verb.Word.Lemma == "score" || verb.Word.Lemma == "have") && verb.Dependents != nil {
			for _, phrase := range sourcePhrases {
				w := t.FindWordByTxt(*verb.Dependents, phrase.Word.Text)
				if w != nil {
					proposerSourcePhrase = phrase
					break
				}
			}

			if proposerSourcePhrase != nil {
				break
			}
		}
	}

	for _, p := range nounPhrases {
		if p.Source != nil && p.Source != proposerSourcePhrase.Source {
			recipientSourcePhrase = p
			break
		}
	}
	fmt.Println("proposer source ", *proposerSourcePhrase.Source.Name, " recipient source ", *recipientSourcePhrase.Source.Name)

	// Scrape Data

	// scraper.ScrapeSources()
	// allGames := scraper.ScrapeThisWeeksGames()

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
