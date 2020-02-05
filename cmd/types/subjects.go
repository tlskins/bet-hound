package types

import "time"

type Subject interface {
	isSubject()
}

type SubjectUnion interface {
	isSubjectUnion()
}

// Player

type Player struct {
	Id        string     `bson:"_id" json:"id"`
	LeagueId  string     `bson:"lg_id,omitempty" json:"league_id"`
	Fk        string     `bson:"fk,omitempty" json:"fk"`
	Name      string     `bson:"name,omitempty" json:"name"`
	Url       string     `bson:"url,omitempty" json:"url"`
	UpdatedAt *time.Time `bson:"upd,omitempty" json:"updated_at"`
	Game      *Game      `bson:"gm,omitempty" json:"game,omitempty"` // should only be aggregated

	FirstName string `bson:"f_name,omitempty" json:"first_name"`
	LastName  string `bson:"l_name,omitempty" json:"last_name"`
	TeamFk    string `bson:"team_fk,omitempty" json:"team_fk"`
	TeamName  string `bson:"team_name,omitempty" json:"team_name"`
	Position  string `bson:"pos,omitempty" json:"position"`
}

func (p Player) isSubject()      {}
func (p Player) isSubjectUnion() {}

// Team

type Team struct {
	Id        string     `bson:"_id" json:"id"`
	LeagueId  string     `bson:"lg_id,omitempty" json:"league_id"`
	Fk        string     `bson:"fk,omitempty" json:"fk"`
	Name      string     `bson:"nm,omitempty" json:"name"`
	Url       string     `bson:"url,omitempty" json:"url"`
	UpdatedAt *time.Time `bson:"upd,omitempty" json:"updated_at"`
	Game      *Game      `bson:"gm,omitempty" json:"game,omitempty"` // should only be aggregated

	Location string `bson:"loc,omitempty" json:"location"`
}

func (t Team) isSubject()      {}
func (t Team) isSubjectUnion() {}
