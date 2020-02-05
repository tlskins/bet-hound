package types

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
