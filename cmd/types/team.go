package types

// Team

type Team struct {
	Id        string `bson:"_id" json:"id"`
	LeagueId  string `bson:"lg_id,omitempty" json:"league_id"`
	Fk        string `bson:"fk,omitempty" json:"fk"`
	Url       string `bson:"url,omitempty" json:"url"`
	Name      string `bson:"nm,omitempty" json:"name"`
	ShortName string `bson:"sht_nm,omitempty" json:"short_name"`
	Location  string `bson:"loc,omitempty" json:"location"`
	Game      *Game  `bson:"gm,omitempty" json:"game,omitempty"` // should only be aggregated
}
