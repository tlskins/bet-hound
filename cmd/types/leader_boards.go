package types

import (
	"time"
)

type LeaderBoard struct {
	Id        string    `bson:"_id" json:"id"`
	LeagueId  string    `bson:"lg_id" json:"league_id"`
	StartTime time.Time `bson:"st" json:"start_time"`
	EndTime   time.Time `bson:"end" json:"end_time"`
	Leaders   []Leader  `bson:"ldrs" json:"leaders"`
	Final     bool      `bson:"fin" json:"final"`
}

type Leader struct {
	UserId   string   `bson:"usr_id" json:"user_id"`
	Rank     int      `bson:"rk" json:"rank"`
	Score    float64  `bson:"scr" json:"score"`
	Wins     int      `bson:"ws" json:"wins"`
	Losses   int      `bson:"ls" json:"losses"`
	WonBets  []string `bson:"w_bets" json:"won_bets"`
	LostBets []string `bson:"l_bets" json:"lost_bets"`
}
