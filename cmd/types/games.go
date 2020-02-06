package types

import (
	"reflect"
	"time"
)

type GamesAggregateInt struct {
	Value int `bson:"value"`
}

type GamesAggregateTime struct {
	Value time.Time `bson:"value"`
}

type GameAndLog struct {
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
}

func (g Game) VsTeamFk(playerTmFk string) string {
	if g.AwayTeamFk == playerTmFk {
		return g.HomeTeamName
	} else if g.HomeTeamFk == playerTmFk {
		return g.AwayTeamName
	}
	return ""
}

// Logs

type GameLog struct {
	HomeTeamLog TeamLog               `bson:"h_tm_lg" json:"home_team_log"`
	AwayTeamLog TeamLog               `bson:"a_tm_lg" json:"away_team_log"`
	PlayerLogs  map[string]*PlayerLog `bson:"p_logs" json:"player_logs"`
}

func (g *GameLog) EvaluateWinner() {
	if g.HomeTeamLog.Score == g.AwayTeamLog.Score {
		g.HomeTeamLog.Win = 0
		g.AwayTeamLog.Win = 0
	} else if g.HomeTeamLog.Score > g.AwayTeamLog.Score {
		g.HomeTeamLog.Win = 1
		g.AwayTeamLog.Win = -1
	} else {
		g.HomeTeamLog.Win = -1
		g.AwayTeamLog.Win = 1
	}
	g.HomeTeamLog.WinBy = g.HomeTeamLog.Score - g.AwayTeamLog.Score
	g.HomeTeamLog.LoseBy = -1 * g.HomeTeamLog.WinBy
	g.AwayTeamLog.WinBy = g.AwayTeamLog.Score - g.HomeTeamLog.Score
	g.AwayTeamLog.LoseBy = -1 * g.AwayTeamLog.WinBy

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
	Fk       string  `bson:"fk" json:"fk"`
	TeamName string  `bson:"tm_nm" json:"team_name"`
	Score    float64 `bson:"scr" json:"score"`
	Win      float64 `bson:"w" json:"win"` // -1 lose, 0 tie, 1 win
	WinBy    float64 `bson:"w_by" json:"win_by"`
	LoseBy   float64 `bson:"l_by" json:"lose_by"`
}

func (t TeamLog) EvaluateMetric(metricField string) *float64 {
	return EvaluateLogMetric(t, metricField)
}

type PlayerLog interface {
	EvaluateMetric(string) *float64
}

// Helpers

func EvaluateLogMetric(log interface{}, metricField string) *float64 {
	if log == nil {
		zero := 0.0
		return &zero
	}
	r := reflect.ValueOf(log)
	r = r.Elem()
	v := r.FieldByName(metricField)
	value := v.Float()
	return &value
}
