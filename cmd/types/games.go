package types

import (
	"math"
	"time"
)

type GamesAggregateInt struct {
	Value int `bson:"value"`
}

type GamesAggregateTime struct {
	Value time.Time `bson:"value"`
}

type Game struct {
	Id            string     `bson:"_id,omitempty" json:"id"`
	LeagueId      string     `bson:"lg_id,omitempty" json:"league_id"`
	Name          string     `bson:"name,omitempty" json:"name"`
	Fk            string     `bson:"fk,omitempty" json:"fk"`
	Url           string     `bson:"url,omitempty" json:"url"`
	AwayTeamFk    string     `bson:"a_team_fk,omitempty" json:"away_team_fk"`
	AwayTeamName  string     `bson:"a_team_name,omitempty" json:"away_team_name"`
	HomeTeamFk    string     `bson:"h_team_fk,omitempty" json:"home_team_fk"`
	HomeTeamName  string     `bson:"h_team_name,omitempty" json:"home_team_name"`
	GameTime      time.Time  `bson:"gm_time,omitempty" json:"game_time"`
	GameResultsAt time.Time  `bson:"gm_res_at,omitempty" json:"game_results_at"`
	UpdatedAt     *time.Time `bson:"upd,omitempty" json:"updated_at"`
	GameLog       *GameLog   `bson:"log,omitempty" json:"game_log"`
}

func (g Game) VsTeamFk(playerTmFk string) string {
	if g.AwayTeamFk == playerTmFk {
		return g.HomeTeamName
	} else if g.HomeTeamFk == playerTmFk {
		return g.AwayTeamName
	}
	return ""
}

type GameLog struct {
	HomeTeamLog TeamLog               `bson:"h_tm_lg" json:"home_team_log"`
	AwayTeamLog TeamLog               `bson:"a_tm_lg" json:"away_team_log"`
	PlayerLogs  map[string]*PlayerLog `bson:"p_logs" json:"player_logs"`
}

func (g GameLog) TeamLogFor(fk string) *TeamLog {
	if g.HomeTeamLog.Fk == fk {
		return &g.HomeTeamLog
	} else if g.AwayTeamLog.Fk == fk {
		return &g.AwayTeamLog
	}
	return nil
}

type TeamLog struct {
	Fk         string    `bson:"fk" json:"fk"`
	TeamName   string    `bson:"tm_nm" json:"team_name"`
	Score      float64   `bson:"scr" json:"score"`
	ScoreByQtr []float64 `bson:"scr_q" json:"score_by_qtr"`
	Win        float64   `bson:"w" json:"win"` // -1 lose, 0 tie, 1 win
	WinBy      float64   `bson:"w_by" json:"win_by"`
	LoseBy     float64   `bson:"l_by" json:"lose_by"`
}

type PlayerLog struct {
	PassCmp      float64 `bson:"p_cmp" json:"pass_cmp"`
	PassAtt      float64 `bson:"p_att" json:"pass_att"`
	PassYd       float64 `bson:"p_yd" json:"pass_yd"`
	PassTd       float64 `bson:"p_td" json:"pass_td"`
	PassInt      float64 `bson:"p_float64" json:"pass_int"`
	PassSacked   float64 `bson:"p_skd" json:"pass_sacked"`
	PassSackedYd float64 `bson:"p_skd_yd" json:"pass_sacked_yd"`
	PassLong     float64 `bson:"p_lng" json:"pass_long"`
	PassRating   float64 `bson:"p_rtg" json:"pass_rating"`
	RushAtt      float64 `bson:"r_att" json:"rush_att"`
	RushYd       float64 `bson:"r_yd" json:"rush_yd"`
	RushTd       float64 `bson:"r_td" json:"rush_td"`
	RushLong     float64 `bson:"r_lng" json:"rush_long"`
	Target       float64 `bson:"tgt" json:"target"`
	Rec          float64 `bson:"rec" json:"rec"`
	RecYd        float64 `bson:"rec_yd" json:"rec_yd"`
	RecTd        float64 `bson:"rec_td" json:"rec_td"`
	RecLong      float64 `bson:"rec_lng" json:"rec_long"`
	Fumble       float64 `bson:"fmbl" json:"fumble"`
	FumbleLost   float64 `bson:"fmbl_lst" json:"fumble_lost"`
	Fantasy00PPR float64 `bson:"f_00_ppr" json:"fantasy_00_ppr"`
	Fantasy05PPR float64 `bson:"f_05_ppr" json:"fantasy_05_ppr"`
	Fantasy10PPR float64 `bson:"f_10_ppr" json:"fantasy_10_ppr"`
}

func (s *PlayerLog) CalcFantasyScores() {
	s.Fantasy00PPR = s.calcFantasyScore(0.0)
	s.Fantasy05PPR = s.calcFantasyScore(0.5)
	s.Fantasy10PPR = s.calcFantasyScore(1.0)
}

func (s PlayerLog) calcFantasyScore(ppr float64) float64 {
	score := 0.0
	score += s.PassYd * 0.04
	score += s.PassTd * 4.0
	score -= s.PassInt * 2.0
	score += s.RushYd * 0.1
	score += s.RushTd * 6.0
	score += s.Rec * ppr
	score += s.RecYd * 0.1
	score += s.RecTd * 6.0
	score -= s.FumbleLost * 2.0
	return math.Ceil(score*10) / 10
}
