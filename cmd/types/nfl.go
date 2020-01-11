package types

type LeagueSettings struct {
	Id           string    `bson:"_id,omitempty" json:"id"`
	CurrentYear  int       `bson:"c_yr,omitempty" json:"current_year"`
	CurrentWeek  int       `bson:"c_wk,omitempty" json:"current_week"`
	PlayerBets   []*BetMap `bson:"p_bts,omitempty" json:"player_bets"`
	TeamBets     []*BetMap `bson:"t_bts,omitempty" json:"team_bets"`
	BetEquations []*BetMap `bson:"b_eqs,omitempty" json:"bet_equations"`
}

func (s LeagueSettings) PlayerBetsMap() (betMap map[int]*BetMap) {
	betMap = make(map[int]*BetMap)
	for _, bet := range s.PlayerBets {
		betMap[bet.Id] = bet
	}
	return
}

func (s LeagueSettings) TeamBetsMap() (betMap map[int]*BetMap) {
	betMap = make(map[int]*BetMap)
	for _, bet := range s.TeamBets {
		betMap[bet.Id] = bet
	}
	return
}

func (s LeagueSettings) BetEquationsMap() (betMap map[int]*BetMap) {
	betMap = make(map[int]*BetMap)
	for _, bet := range s.BetEquations {
		betMap[bet.Id] = bet
	}
	return
}
