package types

import "time"

type GameStat struct {
	Id           string  `bson:"_id,omitempty" json:"id"`
	PlayerFk     string  `bson:"player_fk" json:"player_fk"`
	PassCmp      int64   `bson:"p_cmp" json:"pass_cmp"`
	PassAtt      int64   `bson:"p_att" json:"pass_att"`
	PassYd       int64   `bson:"p_yd" json:"pass_yd"`
	PassTd       int64   `bson:"p_td" json:"pass_td"`
	PassInt      int64   `bson:"p_int" json:"pass_int"`
	PassSacked   int64   `bson:"p_skd" json:"pass_sacked"`
	PassSackedYd int64   `bson:"p_skd_yd" json:"pass_sacked_yd"`
	PassLong     int64   `bson:"p_lng" json:"pass_long"`
	PassRating   float64 `bson:"p_rtg" json:"pass_rating"`
	RushAtt      int64   `bson:"r_att" json:"rush_att"`
	RushYd       int64   `bson:"r_yd" json:"rush_yd"`
	RushTd       int64   `bson:"r_td" json:"rush_td"`
	RushLong     int64   `bson:"r_lng" json:"rush_long"`
	Target       int64   `bson:"tgt" json:"target"`
	Rec          int64   `bson:"rec" json:"rec"`
	RecYd        int64   `bson:"rec_yd" json:"rec_yd"`
	RecTd        int64   `bson:"rec_td" json: "rec_td"`
	RecLong      int64   `bson:"rec_lng" json:"rec_long"`
	Fumble       int64   `bson:"fmbl" json:"fumble"`
	FumbleLost   int64   `bson:"fmbl_lst" json:"fumble_lost"`
}

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
}

func (g Game) VsTeamFk(playerTmFk string) string {
	if g.AwayTeamFk == playerTmFk {
		return g.HomeTeamName
	} else if g.HomeTeamFk == playerTmFk {
		return g.AwayTeamName
	}
	return ""
}
