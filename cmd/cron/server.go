package cron

import (
	"context"
	"fmt"
	"log"
	"time"

	crn "github.com/robfig/cron/v3"

	"bet-hound/cmd/betting"
	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/gql"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
)

func Init(logger *log.Logger, gqlConfig *gql.Config) *crn.Cron {
	fmt.Println("Initializing cron server...")
	cronSrv := crn.New(crn.WithLocation(env.TimeZone()))

	if _, err := cronSrv.AddFunc(fmt.Sprintf("CRON_TZ=%s */30 * * * *", env.ServerTz()), ScrapeAndPushRoto(gqlConfig)); err != nil {
		fmt.Println(err)
	}
	if _, err := cronSrv.AddFunc(fmt.Sprintf("CRON_TZ=%s 0 9 * * *", env.ServerTz()), func() {
		CheckNbaGameResults(logger)
		leagues := CheckNbaBetResults(logger)
		UpdateLeaderBoards(logger, leagues)
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
		logInfo(logger, event, fmt.Sprintf("Scraping game log for %s", game.Id))
		scraper.ScrapeNbaGameLog(game)
	}
}

func CheckNbaBetResults(logger *log.Logger) (leagues map[string]bool) {
	event := "Checking NBA Bet Result"
	fmt.Printf("%s...\n", event)
	readyBets, err := db.GetResultReadyBets("nba")
	logError(logger, err, event)
	leagues = make(map[string]bool)
	for _, bet := range readyBets {
		// evalute and persist bet
		if bet, err = evaluateBet(logger, event, bet); err != nil {
			continue
		}
		// sync profiles
		if _, err := db.SyncBetWithUsers("Final", bet); err != nil {
			logError(logger, err, event)
			continue
		}
		// tweet and mutate and persist bet
		twtTxt := "N/A"
		if tweet, _ := tweetBetResult(logger, event, bet); tweet != nil {
			twtTxt = tweet.GetText()
		}
		logInfo(logger, event, fmt.Sprintf(
			"Status: %s\nResult: %s\nTweet: %s\n",
			bet.BetStatus.String(),
			bet.ResultString(),
			twtTxt,
		))
		leagues[bet.LeagueId] = true
	}
	return
}

func UpdateLeaderBoards(logger *log.Logger, leagues map[string]bool) {
	event := "Update Leader Boards"
	fmt.Printf("%s...\n", event)
	for leagueId, _ := range leagues {
		startWk, endWk := currentLeaderBoardWeek()
		if board, err := db.BuildLeaderBoard(startWk, endWk, leagueId); err != nil {
			logError(logger, err, event)
		} else if err := db.UpsertLeaderBoard(board); err != nil {
			logError(logger, err, event)
		} else {
			logInfo(logger, event, fmt.Sprintf("Updated Leaderboard: %s\n", board.Id))
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

func evaluateBet(logger *log.Logger, event string, bet *t.Bet) (*t.Bet, error) {
	logInfo(logger, event, fmt.Sprintf("Processing bet %s", bet.Id))
	evalBet, err := betting.EvaluateBet(bet)
	if err != nil {
		logError(logger, err, event)
		return nil, err
	}
	if err := db.UpsertBet(evalBet); err != nil {
		logError(logger, err, event)
		return nil, err
	}
	return evalBet, nil
}

func tweetBetResult(logger *log.Logger, event string, bet *t.Bet) (*t.Tweet, error) {
	tweet, err := betting.TweetBetResult(bet)
	if err != nil {
		logError(logger, err, event)
		return nil, err
	}
	if err := db.UpsertBet(bet); err != nil {
		logError(logger, err, event)
		return nil, err
	}
	return tweet, nil
}

func currentLeaderBoardWeek() (*time.Time, *time.Time) {
	now := time.Now().In(env.TimeZone())
	startWk := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, env.TimeZone())
	if wd := startWk.Weekday(); wd == time.Sunday {
		startWk = startWk.AddDate(0, 0, -6)
	} else {
		startWk = startWk.AddDate(0, 0, -int(wd)+1)
	}
	endWk := startWk.AddDate(0, 0, 7)
	return &startWk, &endWk
}

func logError(logger *log.Logger, err error, event string) {
	if err == nil {
		return
	}
	errTxt := fmt.Sprintf("%s [Error]: %s", event, err.Error())
	logger.Printf(errTxt)
	fmt.Println(errTxt)
}

func logInfo(logger *log.Logger, event, msg string) {
	txt := fmt.Sprintf("%s [Info]: %s", event, msg)
	logger.Printf(txt)
	fmt.Println(txt)
}
