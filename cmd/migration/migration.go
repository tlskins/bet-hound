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
		Id:       1,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Pass Completions",
		Field:    "PassCmp",
	}
	playerBets[1] = &t.BetMap{
		Id:       2,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Pass Attempts",
		Field:    "PassAtt",
	}
	playerBets[2] = &t.BetMap{
		Id:       3,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passing Yards",
		Field:    "PassYd",
	}
	playerBets[3] = &t.BetMap{
		Id:       4,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passing Touchdowns",
		Field:    "PassTd",
	}
	playerBets[4] = &t.BetMap{
		Id:       5,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passing Interceptions",
		Field:    "PassInt",
	}
	playerBets[5] = &t.BetMap{
		Id:       6,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Sacks Taken",
		Field:    "PassSacked",
	}
	playerBets[6] = &t.BetMap{
		Id:       7,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Sack Yards Taken",
		Field:    "PassSackedYd",
	}
	playerBets[7] = &t.BetMap{
		Id:       8,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Longest Pass",
		Field:    "PassLong",
	}
	playerBets[8] = &t.BetMap{
		Id:       9,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passer Rating",
		Field:    "PassRating",
	}
	playerBets[9] = &t.BetMap{
		Id:       10,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Rush Attempts",
		Field:    "RushAtt",
	}
	playerBets[10] = &t.BetMap{
		Id:       11,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Rush Yards",
		Field:    "RushYd",
	}
	playerBets[11] = &t.BetMap{
		Id:       12,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Rushing Touchdowns",
		Field:    "RushTd",
	}
	playerBets[12] = &t.BetMap{
		Id:       13,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Longest Run",
		Field:    "RushLong",
	}
	playerBets[13] = &t.BetMap{
		Id:       14,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passing Targets",
		Field:    "Target",
	}
	playerBets[14] = &t.BetMap{
		Id:       15,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Receptions",
		Field:    "Rec",
	}
	playerBets[15] = &t.BetMap{
		Id:       16,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Reception Yards",
		Field:    "RecYd",
	}
	playerBets[16] = &t.BetMap{
		Id:       17,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Reception Touchdowns",
		Field:    "RecTd",
	}
	playerBets[17] = &t.BetMap{
		Id:       18,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Longest Reception",
		Field:    "RecLong",
	}
	playerBets[18] = &t.BetMap{
		Id:       19,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fumbles",
		Field:    "Fumble",
	}
	playerBets[19] = &t.BetMap{
		Id:       20,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fumbles Lost",
		Field:    "FumbleLost",
	}
	playerBets[20] = &t.BetMap{
		Id:       21,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fantasy Points (0.0 PPR)",
		Field:    "Fantasy00PPR",
	}
	playerBets[21] = &t.BetMap{
		Id:       22,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fantasy Points (0.5 PPR)",
		Field:    "Fantasy05PPR",
	}
	playerBets[22] = &t.BetMap{
		Id:       23,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fantasy Points (1.0 PPR)",
		Field:    "Fantasy10PPR",
	}

	// team metrics
	teamBets := make([]*t.BetMap, 5)
	eqId := 31
	gtId := 29
	static := []string{"Static"}
	one := 1.0
	nOne := -1.0
	teamBets[0] = &t.BetMap{
		Id:                   24,
		LeagueId:             "nfl",
		Type:                 "TeamMetric",
		Name:                 "Win",
		Field:                "Win",
		LeftOnly:             true,
		OperatorId:           &eqId,
		RightExpressionValue: &one,
	}
	teamBets[1] = &t.BetMap{
		Id:                   25,
		LeagueId:             "nfl",
		Type:                 "TeamMetric",
		Name:                 "Win By",
		Field:                "WinBy",
		LeftOnly:             true,
		OperatorId:           &gtId,
		RightExpressionTypes: &static,
	}
	teamBets[2] = &t.BetMap{
		Id:                   26,
		LeagueId:             "nfl",
		Type:                 "TeamMetric",
		Name:                 "Lose",
		Field:                "Win",
		LeftOnly:             true,
		OperatorId:           &eqId,
		RightExpressionValue: &nOne,
	}
	teamBets[3] = &t.BetMap{
		Id:                   27,
		LeagueId:             "nfl",
		Type:                 "TeamMetric",
		Name:                 "Lose By",
		Field:                "LoseBy",
		LeftOnly:             true,
		OperatorId:           &gtId,
		RightExpressionTypes: &static,
	}
	teamBets[4] = &t.BetMap{
		Id:                   28,
		LeagueId:             "nfl",
		Type:                 "TeamMetric",
		Name:                 "Points Scored",
		Field:                "Score",
		LeftOnly:             true,
		OperatorId:           &gtId,
		RightExpressionTypes: &static,
	}

	// equalities
	eqs := make([]*t.BetMap, 3)
	eqs[0] = &t.BetMap{
		Id:       29,
		LeagueId: "*",
		Type:     "Operator",
		Name:     ">",
		Field:    "GreaterThan",
	}
	eqs[1] = &t.BetMap{
		Id:       30,
		LeagueId: "*",
		Type:     "Operator",
		Name:     "<",
		Field:    "LesserThan",
	}
	eqs[2] = &t.BetMap{
		Id:       31,
		LeagueId: "*",
		Type:     "Operator",
		Name:     "=",
		Field:    "EqualTo",
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

func SeedNflPlayers() {
	if err := scraper.ScrapePlayers(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Seeded nfl players...")
}

func SeedNflTeams() {
	if err := scraper.ScrapeNflTeams(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Seeded nfl teams...")
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
