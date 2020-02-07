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

type GameAndLog interface {
	GetGameLog() GameLog
}

type GameLog interface {
	TeamLogFor(string) SubjectLog
	PlayerLogFor(string) SubjectLog
}

type SubjectLog interface {
	EvaluateMetric(string) *float64
}

// Helpers

func EvaluateLogMetric(log interface{}, metricField string) *float64 {
	if log == nil {
		zero := 0.0
		return &zero
	}

	v := reflect.ValueOf(log)
	value := v.FieldByName(metricField)
	result := value.Float()
	return &result
}
