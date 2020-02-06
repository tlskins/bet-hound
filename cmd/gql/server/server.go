package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"math"
// 	"os"
// 	"time"

// 	b "bet-hound/cmd/betting"
// 	"bet-hound/cmd/db"
// 	"bet-hound/cmd/env"
// 	"bet-hound/cmd/gql"
// 	"bet-hound/cmd/scraper"
// 	t "bet-hound/cmd/types"
// )

// func InitLeagueSettings(leagueId string) *t.LeagueSettings {
// 	s, err := db.GetLeagueSettings(leagueId)
// 	if err != nil {
// 		panic(err)
// 	}

// 	s.Timezone, _ = time.LoadLocation(env.ServerTz())
// 	const longForm = "Jan 2, 2006 3:04pm (MST)"
// 	lgStart, _ := time.ParseInLocation(longForm, env.LeagueStart(), s.Timezone)
// 	lgStart2, _ := time.ParseInLocation(longForm, env.LeagueStart2(), s.Timezone)
// 	lgEnd, _ := time.ParseInLocation(longForm, env.LeagueEnd(), s.Timezone)

// 	s.MaxScrapedWeek, err = db.GetGamesCurrentWeek(lgStart.Year())
// 	if err != nil {
// 		panic(err)
// 	}
// 	s.MinGameTime, err = db.GetMinGameResultReadyTime()
// 	if err != nil {
// 		panic(err)
// 	}
// 	s.CurrentWeek = CurrentWeek(&lgStart, &lgStart2, &lgEnd)
// 	if s.CurrentWeek > env.LeagueLastWeek() {
// 		s.CurrentWeek = env.LeagueLastWeek()
// 	}

// 	s.StartDate = &lgStart
// 	s.StartWeekTwo = &lgStart2
// 	s.EndDate = &lgEnd
// 	s.CurrentYear = lgStart.Year()
// 	s.LeagueLastWeek = env.LeagueLastWeek()

// 	return s
// }

// func ProcessEvents(s *t.LeagueSettings, logger *log.Logger) func() {
// 	return func() {
// 		fmt.Printf("Processing events @ %s\n", time.Now().In(s.Timezone).String())
// 		logger.Printf("Processing events @ %s\n", time.Now().In(s.Timezone).String())
// 		if currentWk := CurrentWeek(s.StartDate, s.StartWeekTwo, s.EndDate); currentWk != s.CurrentWeek && currentWk <= s.LeagueLastWeek {
// 			s.CurrentWeek = currentWk
// 			fmt.Println("updated week to: ", currentWk)
// 		}

// 		if err := CheckCurrentGames(s); err != nil {
// 			fmt.Println(err)
// 			logger.Println(err)
// 		}
// 		if games, err := CheckGameResults(s); err != nil || games == nil {
// 			if err != nil {
// 				fmt.Println(err)
// 				logger.Println(err)
// 			}
// 		} else if err = ProcessBets(s, games); err != nil {
// 			fmt.Println(err)
// 			logger.Println(err)
// 		}
// 	}
// }

// func CheckGameResults(s *t.LeagueSettings) (*[]*t.Game, error) {
// 	fmt.Printf("%s: Checking game results...\n", time.Now().String())
// 	if s.MinGameTime == nil || s.CurrentWeek == 0 || s.CurrentYear == 0 || time.Now().Before(*s.MinGameTime) {
// 		return nil, nil
// 	}
// 	fmt.Printf("%s: Evaluating game results...\n", time.Now().String())

// 	games, err := db.GetResultReadyGames()
// 	if err != nil {
// 		return nil, err
// 	}
// 	results := []*t.Game{}
// 	for _, game := range games {
// 		log, err := scraper.ScrapeGameLog(game.Url)
// 		if err != nil {
// 			return nil, err
// 		}
// 		game.GameLog = log
// 		results = append(results, game)
// 	}
// 	db.UpsertGames(&results)

// 	minGmTime, err := db.GetMinGameResultReadyTime()
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	s.MinGameTime = minGmTime
// 	return &results, nil
// }

// func CheckCurrentGames(s *t.LeagueSettings) error {
// 	fmt.Printf("%s: Checking current games...\n", time.Now().String())
// 	if s.CurrentWeek == 0 || s.CurrentYear == 0 || s.CurrentWeek == s.MaxScrapedWeek {
// 		return nil
// 	} else if s.CurrentWeek < s.MaxScrapedWeek {
// 		return fmt.Errorf("Current week < Max scraped week!")
// 	}
// 	fmt.Printf("%s: Evaluating current games...\n", time.Now().String())

// 	if err := scraper.ScrapeGames(s.CurrentYear, s.CurrentWeek); err != nil {
// 		return err
// 	}
// 	if err := scraper.ScrapePlayers(); err != nil {
// 		return err
// 	}

// 	s.MaxScrapedWeek = s.CurrentWeek
// 	return nil
// }

// func CurrentWeek(startDate, startWeekTwo, endDate *time.Time) (wk int) {
// 	now := time.Now()
// 	if now.After(*startDate) && now.Before(*endDate) {
// 		if now.Before(*startWeekTwo) {
// 			wk = 1
// 		} else {
// 			wkDiff := now.Sub(*startWeekTwo).Hours() / (24.0 * 7.0)
// 			wk = int(math.Ceil(wkDiff)) + 1
// 		}
// 	}
// 	return
// }
