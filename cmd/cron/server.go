package cron

import (
	"fmt"
	"log"

	"bet-hound/cmd/db"
	"bet-hound/cmd/scraper"
)

func CheckNbaGameResults(logger *log.Logger) {
	event := "Checking Game Result"
	fmt.Printf("%s...\n", event)
	readyGames, err := db.GetResultReadyGames("nba")
	logError(logger, err, event)
	for _, game := range readyGames {
		logInfo(logger, fmt.Sprintf("Scraping game log for %s", game.Id), event)
		scraper.ScrapeNbaGameLog(game)
	}
}

// helpers

func logError(logger *log.Logger, err error, event string) {
	if err == nil {
		return
	}
	errTxt := fmt.Sprintf("%s [Error]: %s", event, err.Error())
	logger.Printf(errTxt)
	fmt.Println(errTxt)
}

func logInfo(logger *log.Logger, msg, event string) {
	txt := fmt.Sprintf("%s [Info]: %s", event, msg)
	logger.Printf(txt)
	fmt.Println(txt)
}
