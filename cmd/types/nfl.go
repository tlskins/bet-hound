package types

import (
	"math"
)

type NflPlayerLog struct {
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

func (t NflPlayerLog) EvaluateMetric(metricField string) *float64 {
	return EvaluateLogMetric(t, metricField)
}

func (s *NflPlayerLog) CalcFantasyScores() {
	s.Fantasy00PPR = s.calcFantasyScore(0.0)
	s.Fantasy05PPR = s.calcFantasyScore(0.5)
	s.Fantasy10PPR = s.calcFantasyScore(1.0)
}

func (s NflPlayerLog) calcFantasyScore(ppr float64) float64 {
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
