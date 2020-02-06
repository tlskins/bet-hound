package cron

import (
	"fmt"
	"log"

	"bet-hound/cmd/betting"
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

func CheckNbaBetResults(logger *log.Logger) {
	event := "Checking NBA Bet Result"
	fmt.Printf("%s...\n", event)
	readyBets, err := db.GetResultReadyBets("nba")
	logError(logger, err, event)
	for _, bet := range readyBets {
		logInfo(logger, fmt.Sprintf("Processing bet %s", bet.Id), event)
		betting.EvaluateBet(bet)
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
