package cron

import (
	"context"
	"fmt"
	"log"

	crn "github.com/robfig/cron/v3"

	"bet-hound/cmd/betting"
	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/gql"
	"bet-hound/cmd/scraper"
)

func Init(logger *log.Logger, gqlConfig *gql.Config) *crn.Cron {
	fmt.Println("Initializing cron server...")
	cronSrv := crn.New(crn.WithLocation(env.TimeZone()))

	if _, err := cronSrv.AddFunc(fmt.Sprintf("CRON_TZ=%s */30 * * * *", env.ServerTz()), ScrapeAndPushRoto(gqlConfig)); err != nil {
		fmt.Println(err)
	}
	if _, err := cronSrv.AddFunc(fmt.Sprintf("CRON_TZ=%s 0 9 * * *", env.ServerTz()), func() {
		CheckNbaGameResults(logger)
		CheckNbaBetResults(logger)
	}); err != nil {
		fmt.Println(err)
	}

	cronSrv.Start()
	return cronSrv
}

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
		if betResult, err := betting.EvaluateBet(bet); err != nil {
			logError(logger, err, event)
		} else {
			msg := fmt.Sprintf(
				"Status: %s\nResult: %s\n",
				betResult.BetStatus.String(),
				betResult.ResultString(),
			)
			logInfo(logger, msg, event)
		}
	}
}

func ScrapeAndPushRoto(config *gql.Config) func() {
	return func() {
		fmt.Println("Scrape and push roto...")
		r := config.Resolvers.Mutation()
		r.PostRotoArticle(context.Background())
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
