package types

import (
	"math"
	"time"
)

type Game struct {
	Id            string    `bson:"_id,omitempty" json:"id"`
	Name          string    `bson:"name,omitempty" json:"name"`
	Fk            string    `bson:"fk,omitempty" json:"fk"`
	Url           string    `bson:"url,omitempty" json:"url"`
	AwayTeamFk    string    `bson:"a_team_fk,omitempty" json:"away_team_fk"`
	AwayTeamName  string    `bson:"a_team_name,omitempty" json:"away_team_name"`
	HomeTeamFk    string    `bson:"h_team_fk,omitempty" json:"home_team_fk"`
	HomeTeamName  string    `bson:"h_team_name,omitempty" json:"home_team_name"`
	GameTime      time.Time `bson:"gm_time,omitempty" json:"game_time"`
	GameResultsAt time.Time `bson:"gm_res_at,omitempty" json:"game_results_at"`
	Final         bool      `bson:"fin" json:"final"`
	Week          int       `bson:"wk" json:"week"`
	Year          int       `bson:"yr" json:"year"`
	GameLog       *GameLog  `bson:"log" json:"game_log"`
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
	HomeTeamLog TeamLog              `bson:"h_tm_lg" json:"home_team_log"`
	AwayTeamLog TeamLog              `bson:"a_tm_lg" json:"away_team_log"`
	PlayerLogs  map[string]PlayerLog `bson:"p_logs" json:"player_logs"`
}

type TeamLog struct {
	Fk         string `bson:"fk" json:"fk"`
	TeamName   string `bson:"tm_nm" json:"team_name"`
	Score      int    `bson:"scr" json:"score"`
	ScoreByQtr []int  `bson:"scr_q" json:"score_by_qtr"`
	Win        int    `bson:"w" json:"win"` // -1 lose, 0 tie, 1 win
}

type PlayerLog struct {
	PassCmp      int     `bson:"p_cmp" json:"pass_cmp"`
	PassAtt      int     `bson:"p_att" json:"pass_att"`
	PassYd       int     `bson:"p_yd" json:"pass_yd"`
	PassTd       int     `bson:"p_td" json:"pass_td"`
	PassInt      int     `bson:"p_int" json:"pass_int"`
	PassSacked   int     `bson:"p_skd" json:"pass_sacked"`
	PassSackedYd int     `bson:"p_skd_yd" json:"pass_sacked_yd"`
	PassLong     int     `bson:"p_lng" json:"pass_long"`
	PassRating   float64 `bson:"p_rtg" json:"pass_rating"`
	RushAtt      int     `bson:"r_att" json:"rush_att"`
	RushYd       int     `bson:"r_yd" json:"rush_yd"`
	RushTd       int     `bson:"r_td" json:"rush_td"`
	RushLong     int     `bson:"r_lng" json:"rush_long"`
	Target       int     `bson:"tgt" json:"target"`
	Rec          int     `bson:"rec" json:"rec"`
	RecYd        int     `bson:"rec_yd" json:"rec_yd"`
	RecTd        int     `bson:"rec_td" json: "rec_td"`
	RecLong      int     `bson:"rec_lng" json:"rec_long"`
	Fumble       int     `bson:"fmbl" json:"fumble"`
	FumbleLost   int     `bson:"fmbl_lst" json:"fumble_lost"`
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
	score += float64(s.PassYd) * 0.04
	score += float64(s.PassTd) * 4.0
	score -= float64(s.PassInt) * 2.0
	// score -= float64(s.PassSackedYd) / 10.0
	score += float64(s.RushYd) * 0.1
	score += float64(s.RushTd) * 6.0
	score += float64(s.Rec) * ppr
	score += float64(s.RecYd) * 0.1
	score += float64(s.RecTd) * 6.0
	score -= float64(s.FumbleLost) * 2.0
	return math.Ceil(score*10) / 10
}
