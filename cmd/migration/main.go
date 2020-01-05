package main

import (
	"fmt"
	"log"
	"os"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

const appConfigPath = "../env"
const appConfigName = "config"

var logger *log.Logger

func main() {
	// Initialization
	logger = setUpLogger(env.LogPath(), "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading application config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	scraper.ScrapePlayers()

	// Player metrics
	playerBets := make([]t.BetMap, 23)
	playerBets[0] = t.BetMap{
		Name:       "Pass Completions",
		Field:      "PassCmp",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[1] = t.BetMap{
		Name:       "Pass Attempts",
		Field:      "PassAtt",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[2] = t.BetMap{
		Name:       "Passing Yards",
		Field:      "PassYd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[3] = t.BetMap{
		Name:       "Passing Touchdowns",
		Field:      "PassTd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[4] = t.BetMap{
		Name:       "Passing Interceptions",
		Field:      "PassInt",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[5] = t.BetMap{
		Name:       "Sacks Taken",
		Field:      "PassSacked",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[6] = t.BetMap{
		Name:       "Sack Yards Taken",
		Field:      "PassSackedYd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[7] = t.BetMap{
		Name:       "Longest Pass",
		Field:      "PassLong",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[8] = t.BetMap{
		Name:       "Passer Rating",
		Field:      "PassRating",
		FieldType:  "Field",
		ResultType: "float64",
	}
	playerBets[9] = t.BetMap{
		Name:       "Rush Attempts",
		Field:      "RushAtt",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[10] = t.BetMap{
		Name:       "Rush Yards",
		Field:      "RushYd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[11] = t.BetMap{
		Name:       "Rushing Touchdowns",
		Field:      "RushTd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[12] = t.BetMap{
		Name:       "Longest Run",
		Field:      "RushLong",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[13] = t.BetMap{
		Name:       "Passing Targets",
		Field:      "Target",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[14] = t.BetMap{
		Name:       "Receptions",
		Field:      "Rec",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[15] = t.BetMap{
		Name:       "Reception Yards",
		Field:      "RecYd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[16] = t.BetMap{
		Name:       "Reception Touchdowns",
		Field:      "RecTd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[17] = t.BetMap{
		Name:       "Longest Reception",
		Field:      "RecLong",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[18] = t.BetMap{
		Name:       "Fumbles",
		Field:      "Fumble",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[19] = t.BetMap{
		Name:       "Fumbles Lost",
		Field:      "FumbleLost",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[20] = t.BetMap{
		Name:       "Fantasy Points (0.0 PPR)",
		Field:      "Fantasy00PPR",
		FieldType:  "Field",
		ResultType: "float64",
	}
	playerBets[21] = t.BetMap{
		Name:       "Fantasy Points (0.5 PPR)",
		Field:      "Fantasy05PPR",
		FieldType:  "Field",
		ResultType: "float64",
	}
	playerBets[22] = t.BetMap{
		Name:       "Fantasy Points (1.0 PPR)",
		Field:      "Fantasy10PPR",
		FieldType:  "Field",
		ResultType: "float64",
	}

	// team metrics
	teamBets := make([]t.BetMap, 3)
	teamBets[0] = t.BetMap{
		Name:       "Win",
		Field:      "Win",
		FieldType:  "Field",
		ResultType: "int",
	}
	teamBets[1] = t.BetMap{
		Name:       "Lose",
		Field:      "Lose",
		FieldType:  "Func",
		ResultType: "int",
	}
	teamBets[2] = t.BetMap{
		Name:       "Points Scored",
		Field:      "Score",
		FieldType:  "Field",
		ResultType: "int",
	}

	// equalities
	eqs := make([]t.BetMap, 3)
	eqs[0] = t.BetMap{
		Name:       ">",
		Field:      "GreaterThan",
		FieldType:  "Func",
		ResultType: "bool",
	}
	eqs[1] = t.BetMap{
		Name:       "<",
		Field:      "LesserThan",
		FieldType:  "Func",
		ResultType: "bool",
	}
	eqs[2] = t.BetMap{
		Name:       "=",
		Field:      "EqualTo",
		FieldType:  "Func",
		ResultType: "bool",
	}

	setting := t.LeagueSettings{
		Id:           "nfl",
		PlayerBets:   playerBets,
		TeamBets:     teamBets,
		BetEquations: eqs,
	}
	if err := db.UpsertLeagueSettings(&setting); err != nil {
		fmt.Println(err)
	}
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
