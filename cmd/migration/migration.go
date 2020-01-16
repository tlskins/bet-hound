package migration

import (
	"fmt"

	"bet-hound/cmd/db"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
)

func SeedNflLeagueSettings() {
	// Player metrics
	playerBets := make([]*t.BetMap, 23)
	playerBets[0] = &t.BetMap{
		Id:         1,
		Name:       "Pass Completions",
		Field:      "PassCmp",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[1] = &t.BetMap{
		Id:         2,
		Name:       "Pass Attempts",
		Field:      "PassAtt",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[2] = &t.BetMap{
		Id:         3,
		Name:       "Passing Yards",
		Field:      "PassYd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[3] = &t.BetMap{
		Id:         4,
		Name:       "Passing Touchdowns",
		Field:      "PassTd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[4] = &t.BetMap{
		Id:         5,
		Name:       "Passing Interceptions",
		Field:      "PassInt",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[5] = &t.BetMap{
		Id:         6,
		Name:       "Sacks Taken",
		Field:      "PassSacked",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[6] = &t.BetMap{
		Id:         7,
		Name:       "Sack Yards Taken",
		Field:      "PassSackedYd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[7] = &t.BetMap{
		Id:         8,
		Name:       "Longest Pass",
		Field:      "PassLong",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[8] = &t.BetMap{
		Id:         9,
		Name:       "Passer Rating",
		Field:      "PassRating",
		FieldType:  "Field",
		ResultType: "float64",
	}
	playerBets[9] = &t.BetMap{
		Id:         10,
		Name:       "Rush Attempts",
		Field:      "RushAtt",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[10] = &t.BetMap{
		Id:         11,
		Name:       "Rush Yards",
		Field:      "RushYd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[11] = &t.BetMap{
		Id:         12,
		Name:       "Rushing Touchdowns",
		Field:      "RushTd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[12] = &t.BetMap{
		Id:         13,
		Name:       "Longest Run",
		Field:      "RushLong",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[13] = &t.BetMap{
		Id:         14,
		Name:       "Passing Targets",
		Field:      "Target",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[14] = &t.BetMap{
		Id:         15,
		Name:       "Receptions",
		Field:      "Rec",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[15] = &t.BetMap{
		Id:         16,
		Name:       "Reception Yards",
		Field:      "RecYd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[16] = &t.BetMap{
		Id:         17,
		Name:       "Reception Touchdowns",
		Field:      "RecTd",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[17] = &t.BetMap{
		Id:         18,
		Name:       "Longest Reception",
		Field:      "RecLong",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[18] = &t.BetMap{
		Id:         19,
		Name:       "Fumbles",
		Field:      "Fumble",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[19] = &t.BetMap{
		Id:         20,
		Name:       "Fumbles Lost",
		Field:      "FumbleLost",
		FieldType:  "Field",
		ResultType: "int",
	}
	playerBets[20] = &t.BetMap{
		Id:         21,
		Name:       "Fantasy Points (0.0 PPR)",
		Field:      "Fantasy00PPR",
		FieldType:  "Field",
		ResultType: "float64",
	}
	playerBets[21] = &t.BetMap{
		Id:         22,
		Name:       "Fantasy Points (0.5 PPR)",
		Field:      "Fantasy05PPR",
		FieldType:  "Field",
		ResultType: "float64",
	}
	playerBets[22] = &t.BetMap{
		Id:         23,
		Name:       "Fantasy Points (1.0 PPR)",
		Field:      "Fantasy10PPR",
		FieldType:  "Field",
		ResultType: "float64",
	}

	// team metrics
	teamBets := make([]*t.BetMap, 5)
	teamBets[0] = &t.BetMap{
		Id:         24,
		Name:       "Home Team Win",
		Field:      "HomeTeamWin",
		FieldType:  "Func",
		ResultType: "int",
	}
	teamBets[1] = &t.BetMap{
		Id:         25,
		Name:       "Away Team Win",
		Field:      "AwayTeamWin",
		FieldType:  "Func",
		ResultType: "int",
	}
	teamBets[2] = &t.BetMap{
		Id:         26,
		Name:       "Home Team Lose",
		Field:      "AwayTeamWin",
		FieldType:  "Func",
		ResultType: "int",
	}
	teamBets[3] = &t.BetMap{
		Id:         27,
		Name:       "Away Team Lose",
		Field:      "HomeTeamWin",
		FieldType:  "Func",
		ResultType: "int",
	}
	teamBets[4] = &t.BetMap{
		Id:         28,
		Name:       "Points Scored",
		Field:      "Score",
		FieldType:  "Field",
		ResultType: "int",
	}

	// equalities
	eqs := make([]*t.BetMap, 3)
	eqs[0] = &t.BetMap{
		Id:         29,
		Name:       ">",
		Field:      "GreaterThan",
		FieldType:  "Func",
		ResultType: "bool",
	}
	eqs[1] = &t.BetMap{
		Id:         30,
		Name:       "<",
		Field:      "LesserThan",
		FieldType:  "Func",
		ResultType: "bool",
	}
	eqs[2] = &t.BetMap{
		Id:         31,
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
	fmt.Println("Seeded nfl league settings...")
}

func SeedNflPlayers() {
	scraper.ScrapePlayers()
	fmt.Println("Seeded nfl players...")
}

func SeedUsers() {
	tim := t.User{
		Id:       "timlee",
		Name:     "Timothy Lee",
		UserName: "JooSeeDong",
		Password: "password",
		Email:    "tlee87@gmail.com",
		TwitterUser: &t.TwitterUser{
			Id:         501399114,
			ScreenName: "timmy_the_truth",
			Name:       "steve_aioli",
			IdStr:      "501399114",
		},
	}
	xtine := t.User{
		Id:       "xtine",
		Name:     "Christine Kettler",
		UserName: "cktweets",
		Password: "password",
		Email:    "christine.b.kettler@gmail.com",
		TwitterUser: &t.TwitterUser{
			Id:         249778392,
			ScreenName: "ckettstweets",
			Name:       "Christine Kettler",
			IdStr:      "249778392",
		},
	}
	db.UpsertUser(&tim)
	db.UpsertUser(&xtine)
	fmt.Println("Seeded users...")
}
