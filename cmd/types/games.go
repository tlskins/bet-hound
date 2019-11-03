package types

import "time"

type Game struct {
	Id           *string    `bson:"_id,omitempty" json:"id"`
	Name         *string    `bson:"name,omitempty" json:"name"`
	Fk           *string    `bson:"fk,omitempty" json:"fk"`
	Url          *string    `bson:"url,omitempty" json:"url"`
	AwayTeamFk   *string    `bson:"a_team_fk,omitempty" json:"away_team_fk"`
	AwayTeamName *string    `bson:"a_team_name,omitempty" json:"away_team_name"`
	HomeTeamFk   *string    `bson:"h_team_fk,omitempty" json:"home_team_fk"`
	HomeTeamName *string    `bson:"h_team_name,omitempty" json:"home_team_name"`
	GameTime     *time.Time `bson:"gm_time,omitempty" json:"game_time"`
}
