package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	b "bet-hound/cmd/betting"
	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/gql"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
)

func InitLeagueSettings(tz *time.Location) *t.LeagueSettings {
	const longForm = "Jan 2, 2006 3:04pm (MST)"
	// first game sep 5 2019
	lgStart, _ := time.ParseInLocation(longForm, "Sep 2, 2019 9:00am (EDT)", tz)
	lgStart2, _ := time.ParseInLocation(longForm, "Sep 10, 2019 9:00am (EDT)", tz)
	lgEnd, _ := time.ParseInLocation(longForm, "Feb 3, 2020 9:00am (EDT)", tz)

	maxWk, err := db.GetGamesCurrentWeek(lgStart.Year())
	if err != nil {
		panic(err)
	}
	minGmTime, err := db.GetMinGameResultReadyTime()
	if err != nil {
		panic(err)
	}

	s := t.LeagueSettings{
		StartDate:      &lgStart,
		StartWeekTwo:   &lgStart2,
		EndDate:        &lgEnd,
		MaxScrapedWeek: maxWk,
		MinGameTime:    minGmTime,
		CurrentYear:    lgStart.Year(),
		Timezone:       tz,
	}
	s.CurrentWeek = currentWeek(&s)

	return &s
}

func SetUpLogger(logPath, defaultPath string) *log.Logger {
	if logPath == "" {
		logPath = defaultPath
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(f, "", 0)
}

func ProcessRotoNfl(config *gql.Config) func() {
	return func() {
		r := config.Resolvers.Mutation()
		r.PostRotoArticle(context.Background())
	}
}

func ProcessEvents(s *t.LeagueSettings, logger *log.Logger) func() {
	return func() {
		fmt.Printf("Processing events @ %s\n", time.Now().In(s.Timezone).String())
		logger.Printf("Processing events @ %s\n", time.Now().In(s.Timezone).String())
		if err := CheckCurrentGames(s); err != nil {
			logger.Println(err)
		}
		if games, err := CheckGameResults(s); err != nil || games == nil {
			logger.Println(err)
		} else {
			if err = ProcessBets(s, games); err != nil {
				logger.Println(err)
			}
		}
	}
}

func ProcessBets(s *t.LeagueSettings, games *[]*t.Game) error {
	for _, game := range *games {
		bets, err := db.FindAcceptedBetsByGame(game.Id)
		if err != nil {
			return err
		}
		client := env.TwitterClient()
		for _, bet := range *bets {
			if err := bet.Valid(); err != nil {
				fmt.Println("skipping invalid bet ", bet.Id)
				continue
			}
			evBet, err := b.EvaluateBet(bet, game)
			if err != nil {
				return nil
			}

			if evBet.BetStatus.String() == "Final" {
				if evBet.TwitterHandles() != "" && evBet.AcceptFk != "" {
					txt := fmt.Sprintf("%s Congrats %s you beat %s! %s",
						evBet.TwitterHandles(),
						evBet.BetResult.Winner.Name,
						evBet.BetResult.Loser.Name,
						evBet.BetResult.Response,
					)
					resp, err := client.SendTweet(txt, &evBet.AcceptFk)
					if err != nil {
						return err
					}
					evBet.BetResult.ResponseFk = resp.IdStr
				}
			}
			db.UpsertBet(evBet)
		}
	}
	return nil
}

func CheckGameResults(s *t.LeagueSettings) (*[]*t.Game, error) {
	fmt.Printf("%s: Checking game results...\n", time.Now().String())
	if s.MinGameTime == nil || s.CurrentWeek == 0 || s.CurrentYear == 0 || time.Now().Before(*s.MinGameTime) {
		return nil, nil
	}
	fmt.Printf("%s: Evaluating game results...\n", time.Now().String())

	games, err := db.GetResultReadyGames()
	if err != nil {
		return nil, err
	}
	results := []*t.Game{}
	for _, game := range games {
		log, err := scraper.ScrapeGameLog(game.Url)
		if err != nil {
			return nil, err
		}
		game.GameLog = log
		results = append(results, game)
	}
	db.UpsertGames(&results)

	minGmTime, err := db.GetMinGameResultReadyTime()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// s.Mu.Lock()
	s.MinGameTime = minGmTime
	// s.Mu.Unlock()
	return &results, nil
}

func CheckCurrentGames(s *t.LeagueSettings) error {
	fmt.Printf("%s: Checking current games...\n", time.Now().String())
	if s.CurrentWeek == 0 || s.CurrentYear == 0 || s.CurrentWeek == s.MaxScrapedWeek {
		return nil
	} else if s.CurrentWeek < s.MaxScrapedWeek {
		return fmt.Errorf("Current week < Max scraped week!")
	}
	fmt.Printf("%s: Evaluating current games...\n", time.Now().String())

	if err := scraper.ScrapeGames(s.CurrentYear, s.CurrentWeek); err != nil {
		return err
	}

	// s.Mu.Lock()
	s.MaxScrapedWeek = s.CurrentWeek
	// s.Mu.Unlock()
	return nil
}

func currentWeek(s *t.LeagueSettings) (wk int) {
	now := time.Now()
	if now.After(*s.StartDate) && now.Before(*s.EndDate) {
		if now.Before(*s.StartWeekTwo) {
			wk = 1
		} else {
			wkDiff := now.Sub(*s.StartWeekTwo).Hours() / (24.0 * 7.0)
			wk = int(math.Ceil(wkDiff)) + 1
		}
	}
	return
}
