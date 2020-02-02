package types

import (
	"fmt"
	"time"
	// "sync"

	h "bet-hound/pkg/helpers"
)

type LeagueSettings struct {
	Id             string     `bson:"_id" json:"id"`
	StartDate      *time.Time `bson:"lg_st,omitempty" json:"league_start_date"`
	StartWeekTwo   *time.Time `bson:"lg_st_wk_two,omitempty" json:"league_start_week_two"`
	EndDate        *time.Time `bson:"lg_end,omitempty" json:"league_end_date"`
	MaxScrapedWeek int        `bson:"mx_wk,omitempty" json:"max_scraped_week"`
	LeagueLastWeek int        `bson:"lg_lst,omitempty" json:"league_last_week"`
	MinGameTime    *time.Time `bson:"min_gm,omitempty" json:"min_game_time"`
	CurrentYear    int        `bson:"c_yr,omitempty" json:"current_year"`
	CurrentWeek    int        `bson:"c_wk,omitempty" json:"current_week"`
	PlayerBets     []*BetMap  `bson:"p_bts,omitempty" json:"player_bets"`
	TeamBets       []*BetMap  `bson:"t_bts,omitempty" json:"team_bets"`
	BetEquations   []*BetMap  `bson:"b_eqs,omitempty" json:"bet_equations"`
	Timezone       *time.Location
	// Mu             sync.Mutex
}

func (s *LeagueSettings) Print() {
	fmt.Println(h.PrettyPrint(*s))
}

func (s *LeagueSettings) Metrics() (betMap map[int]*BetMap) {
	betMap = make(map[int]*BetMap)
	for _, bet := range s.PlayerBets {
		betMap[bet.Id] = bet
	}
	for _, bet := range s.TeamBets {
		betMap[bet.Id] = bet
	}
	return
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
