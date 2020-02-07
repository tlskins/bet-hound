package types

import (
	"time"
)

type NbaGameAndLog struct {
	Id            string      `bson:"_id,omitempty" json:"id"`
	LeagueId      string      `bson:"lg_id,omitempty" json:"league_id"`
	Name          string      `bson:"name,omitempty" json:"name"`
	Fk            string      `bson:"fk,omitempty" json:"fk"`
	Url           string      `bson:"url,omitempty" json:"url"`
	AwayTeamFk    string      `bson:"a_team_fk,omitempty" json:"away_team_fk"`
	AwayTeamName  string      `bson:"a_team_name,omitempty" json:"away_team_name"`
	HomeTeamFk    string      `bson:"h_team_fk,omitempty" json:"home_team_fk"`
	HomeTeamName  string      `bson:"h_team_name,omitempty" json:"home_team_name"`
	GameTime      time.Time   `bson:"gm_time,omitempty" json:"game_time"`
	GameResultsAt time.Time   `bson:"gm_res_at,omitempty" json:"game_results_at"`
	UpdatedAt     *time.Time  `bson:"upd,omitempty" json:"updated_at"`
	GameLog       *NbaGameLog `bson:"log,omitempty" json:"game_log"`
}

func (g NbaGameAndLog) GetGameLog() GameLog {
	var res GameLog = *g.GameLog
	return res
}

type NbaGameLog struct {
	HomeTeamLog NbaTeamLog               `bson:"h_tm_lg" json:"home_team_log"`
	AwayTeamLog NbaTeamLog               `bson:"a_tm_lg" json:"away_team_log"`
	PlayerLogs  map[string]*NbaPlayerLog `bson:"p_logs" json:"player_logs"`
}

func (g NbaGameLog) TeamLogFor(fk string) SubjectLog {
	var tmLog SubjectLog
	if g.HomeTeamLog.Fk == fk {
		tmLog = g.HomeTeamLog
	} else if g.AwayTeamLog.Fk == fk {
		tmLog = g.AwayTeamLog
	}
	return tmLog
}

func (g NbaGameLog) PlayerLogFor(fk string) SubjectLog {
	var pLog SubjectLog = g.PlayerLogs[fk]
	return pLog
}

type NbaTeamLog struct {
	Fk       string  `bson:"fk" json:"fk"`
	TeamName string  `bson:"tm_nm" json:"team_name"`
	Score    float64 `bson:"scr" json:"score"`
	Win      float64 `bson:"w" json:"win"` // -1 lose, 0 tie, 1 win
	WinBy    float64 `bson:"w_by" json:"win_by"`
	LoseBy   float64 `bson:"l_by" json:"lose_by"`
}

func (t NbaTeamLog) EvaluateMetric(metricField string) *float64 {
	return EvaluateLogMetric(t, metricField)
}

type NbaPlayerLog struct {
	MinsPlayed     float64 `bson:"mins" json:"mins_played"`
	FieldGoals     float64 `bson:"fg" json:"field_goals"`
	FieldGoalAtts  float64 `bson:"fga" json:"field_goal_atts"`
	FieldGoalPct   float64 `bson:"fgp" json:"field_goal_pct"`
	FieldGoal3s    float64 `bson:"fg3" json:"field_goal_3s"`
	FieldGoal3Atts float64 `bson:"fg3a" json:"field_goal_3_atts"`
	FieldGoal3Pct  float64 `bson:"fg3p" json:"field_goal_3_pct"`
	FreeThrows     float64 `bson:"ft" json:"free_throws"`
	FreeThrowAtts  float64 `bson:"fta" json:"free_throw_atts"`
	FreeThrowPct   float64 `bson:"ftp" json:"free_throw_pct"`
	OffRebound     float64 `bson:"oreb" json:"off_rebound"`
	DefRebound     float64 `bson:"dreb" json:"def_rebound"`
	TotalRebounds  float64 `bson:"treb" json:"total_rebounds"`
	Assists        float64 `bson:"ast" json:"assists"`
	Steals         float64 `bson:"stl" json:"steals"`
	Blocks         float64 `bson:"blk" json:"blocks"`
	TurnOvers      float64 `bson:"tov" json:"turnovers"`
	PersonalFouls  float64 `bson:"pfs" json:"personal_fouls"`
	Points         float64 `bson:"pts" json:"points"`
	PlusMinus      float64 `bson:"p_m" json:"plus_minus"`
}

func (t NbaPlayerLog) EvaluateMetric(metricField string) *float64 {
	return EvaluateLogMetric(t, metricField)
}
