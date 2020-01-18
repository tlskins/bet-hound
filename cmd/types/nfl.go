package types

import (
	"fmt"
	// "sync"
	"time"
)

type LeagueSettings struct {
	Id             string     `bson:"_id" json:"id"`
	StartDate      *time.Time `bson:"lg_st" json:"league_start_date"`
	StartWeekTwo   *time.Time `bson:"lg_st_wk_two" json:"league_start_week_two"`
	EndDate        *time.Time `bson:"lg_end" json:"league_end_date"`
	MaxScrapedWeek int        `bson:"mx_wk" json:"max_scraped_week"`
	MinGameTime    *time.Time `bson:"min_gm" json:"min_game_time"`
	CurrentYear    int        `bson:"c_yr" json:"current_year"`
	CurrentWeek    int        `bson:"c_wk" json:"current_week"`
	PlayerBets     []*BetMap  `bson:"p_bts" json:"player_bets"`
	TeamBets       []*BetMap  `bson:"t_bts" json:"team_bets"`
	BetEquations   []*BetMap  `bson:"b_eqs" json:"bet_equations"`
	Timezone       *time.Location
	// Mu             sync.Mutex
}

func (s *LeagueSettings) Print() {
	fmt.Println("StartDate: ", s.StartDate.String())
	fmt.Println("StartWeekTwo: ", s.StartWeekTwo.String())
	fmt.Println("EndDate: ", s.EndDate.String())
	fmt.Println("MaxScrapedWeek: ", s.MaxScrapedWeek)
	fmt.Println("MinGameTime: ", s.MinGameTime.String())
	fmt.Println("CurrentWeek: ", s.CurrentWeek)
	fmt.Println("Timezone: ", s.Timezone.String())
}

func (s *LeagueSettings) PlayerBetsMap() (betMap map[int]*BetMap) {
	betMap = make(map[int]*BetMap)
	for _, bet := range s.PlayerBets {
		betMap[bet.Id] = bet
	}
	return
}

func (s *LeagueSettings) TeamBetsMap() (betMap map[int]*BetMap) {
	betMap = make(map[int]*BetMap)
	for _, bet := range s.TeamBets {
		betMap[bet.Id] = bet
	}
	return
}

func (s *LeagueSettings) BetEquationsMap() (betMap map[int]*BetMap) {
	betMap = make(map[int]*BetMap)
	for _, bet := range s.BetEquations {
		betMap[bet.Id] = bet
	}
	return
}
