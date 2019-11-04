package types

type Source struct {
	Id        *string `bson:"_id,omitempty" json:"id"`
	Name      *string `bson:"name,omitempty" json:"name"`
	FirstName *string `bson:"f_name,omitempty" json:"first_name"`
	LastName  *string `bson:"l_name,omitempty" json:"last_name"`
	Fk        *string `bson:"fk,omitempty" json:"fk"`
	TeamFk    *string `bson:"team_fk,omitempty" json:"team_fk"`
	TeamName  *string `bson:"team_name,omitempty" json:"team_name"`
	Position  *string `bson:"pos,omitempty" json:"position"`
	Url       *string `bson:"url,omitempty" json:"url"`
}
