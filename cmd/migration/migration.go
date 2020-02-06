package migration

import (
	"fmt"

	"bet-hound/cmd/db"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
)

func SeedBetMaps() {
	// Player metrics
	nflPlayerBets := make([]*t.BetMap, 23)
	nflPlayerBets[0] = &t.BetMap{
		Id:       1,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Pass Completions",
		Field:    "PassCmp",
	}
	nflPlayerBets[1] = &t.BetMap{
		Id:       2,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Pass Attempts",
		Field:    "PassAtt",
	}
	nflPlayerBets[2] = &t.BetMap{
		Id:       3,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passing Yards",
		Field:    "PassYd",
	}
	nflPlayerBets[3] = &t.BetMap{
		Id:       4,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passing Touchdowns",
		Field:    "PassTd",
	}
	nflPlayerBets[4] = &t.BetMap{
		Id:       5,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passing Interceptions",
		Field:    "PassInt",
	}
	nflPlayerBets[5] = &t.BetMap{
		Id:       6,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Sacks Taken",
		Field:    "PassSacked",
	}
	nflPlayerBets[6] = &t.BetMap{
		Id:       7,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Sack Yards Taken",
		Field:    "PassSackedYd",
	}
	nflPlayerBets[7] = &t.BetMap{
		Id:       8,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Longest Pass",
		Field:    "PassLong",
	}
	nflPlayerBets[8] = &t.BetMap{
		Id:       9,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passer Rating",
		Field:    "PassRating",
	}
	nflPlayerBets[9] = &t.BetMap{
		Id:       10,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Rush Attempts",
		Field:    "RushAtt",
	}
	nflPlayerBets[10] = &t.BetMap{
		Id:       11,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Rush Yards",
		Field:    "RushYd",
	}
	nflPlayerBets[11] = &t.BetMap{
		Id:       12,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Rushing Touchdowns",
		Field:    "RushTd",
	}
	nflPlayerBets[12] = &t.BetMap{
		Id:       13,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Longest Run",
		Field:    "RushLong",
	}
	nflPlayerBets[13] = &t.BetMap{
		Id:       14,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Passing Targets",
		Field:    "Target",
	}
	nflPlayerBets[14] = &t.BetMap{
		Id:       15,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Receptions",
		Field:    "Rec",
	}
	nflPlayerBets[15] = &t.BetMap{
		Id:       16,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Reception Yards",
		Field:    "RecYd",
	}
	nflPlayerBets[16] = &t.BetMap{
		Id:       17,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Reception Touchdowns",
		Field:    "RecTd",
	}
	nflPlayerBets[17] = &t.BetMap{
		Id:       18,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Longest Reception",
		Field:    "RecLong",
	}
	nflPlayerBets[18] = &t.BetMap{
		Id:       19,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fumbles",
		Field:    "Fumble",
	}
	nflPlayerBets[19] = &t.BetMap{
		Id:       20,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fumbles Lost",
		Field:    "FumbleLost",
	}
	nflPlayerBets[20] = &t.BetMap{
		Id:       21,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fantasy Points (0.0 PPR)",
		Field:    "Fantasy00PPR",
	}
	nflPlayerBets[21] = &t.BetMap{
		Id:       22,
		LeagueId: "nfl",
		Type:     "PlayerMetric",
		Name:     "Fantasy Points (0.5 PPR)",
		Field:    "Fantasy05PPR",
	}
	nflPlayerBets[22] = &t.BetMap{
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

	// Nba Player metrics
	nbaPlayerBets := make([]*t.BetMap, 20)
	nbaPlayerBets[0] = &t.BetMap{
		Id:       32,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Minutes Played",
		Field:    "MinsPlayed",
	}
	nbaPlayerBets[1] = &t.BetMap{
		Id:       33,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Field Goals",
		Field:    "FieldGoals",
	}
	nbaPlayerBets[2] = &t.BetMap{
		Id:       34,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Field Goal Attempts",
		Field:    "FieldGoalAtts",
	}
	nbaPlayerBets[3] = &t.BetMap{
		Id:       35,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Field Goal Percent",
		Field:    "FieldGoalPct",
	}
	nbaPlayerBets[4] = &t.BetMap{
		Id:       36,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Field Goal 3s",
		Field:    "FieldGoal3s",
	}
	nbaPlayerBets[5] = &t.BetMap{
		Id:       37,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Field Goal 3 Attempts",
		Field:    "FieldGoal3Atts",
	}
	nbaPlayerBets[6] = &t.BetMap{
		Id:       38,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Field Goal 3 Percentage",
		Field:    "FieldGoal3Pct",
	}
	nbaPlayerBets[7] = &t.BetMap{
		Id:       39,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Free Throws",
		Field:    "FreeThrows",
	}
	nbaPlayerBets[8] = &t.BetMap{
		Id:       40,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Free Throw Attempts",
		Field:    "FreeThrowAtts",
	}
	nbaPlayerBets[9] = &t.BetMap{
		Id:       41,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Free Throw Percentage",
		Field:    "FreeThrowPct",
	}
	nbaPlayerBets[10] = &t.BetMap{
		Id:       42,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Offensive Rebounds",
		Field:    "OffRebound",
	}
	nbaPlayerBets[11] = &t.BetMap{
		Id:       43,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Defensive Rebounds",
		Field:    "DefRebound",
	}
	nbaPlayerBets[12] = &t.BetMap{
		Id:       44,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Total Rebounds",
		Field:    "TotalRebounds",
	}
	nbaPlayerBets[13] = &t.BetMap{
		Id:       45,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Assists",
		Field:    "Assists",
	}
	nbaPlayerBets[14] = &t.BetMap{
		Id:       46,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Steals",
		Field:    "Steals",
	}
	nbaPlayerBets[15] = &t.BetMap{
		Id:       47,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Blocks",
		Field:    "Blocks",
	}
	nbaPlayerBets[16] = &t.BetMap{
		Id:       48,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "TurnOvers",
		Field:    "TurnOvers",
	}
	nbaPlayerBets[17] = &t.BetMap{
		Id:       49,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Personal Fouls",
		Field:    "PersonalFouls",
	}
	nbaPlayerBets[18] = &t.BetMap{
		Id:       50,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Points",
		Field:    "Points",
	}
	nbaPlayerBets[19] = &t.BetMap{
		Id:       51,
		LeagueId: "nba",
		Type:     "PlayerMetric",
		Name:     "Plus Minus",
		Field:    "PlusMinus",
	}

	if err := db.UpsertBetMaps(&nflPlayerBets); err != nil {
		fmt.Println(err)
	}
	if err := db.UpsertBetMaps(&teamBets); err != nil {
		fmt.Println(err)
	}
	if err := db.UpsertBetMaps(&eqs); err != nil {
		fmt.Println(err)
	}
	if err := db.UpsertBetMaps(&nbaPlayerBets); err != nil {
		fmt.Println(err)
	}
}

func SeedNflPlayers() {
	if err := scraper.ScrapeNflPlayers(); err != nil {
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
