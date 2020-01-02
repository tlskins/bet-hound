package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func UpsertPlayers(players *[]*t.Player) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	for _, player := range *players {
		err = m.Upsert(c, nil, m.M{"_id": player.Id}, m.M{"$set": player})
		if err != nil {
			return err
		}
	}
	return err
}

func SearchPlayers(name, team, position *string, numResults int) (players []*t.Player, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	// TODO : set indexes somewhere else
	index := mgo.Index{Key: []string{"$text:f_name", "$text:l_name", "$text:name"}}
	m.CreateIndex(c, index)

	query := m.M{}
	if name != nil {
		// query["$text"] = m.M{"$search": *name}
		query["name"] = bson.RegEx{*name, "i"}
	}
	if team != nil {
		teamSrch := *team + "*"
		query["$or"] = []m.M{
			m.M{"team_fk": bson.RegEx{teamSrch, "i"}},
			m.M{"team_name": bson.RegEx{teamSrch, "i"}},
			m.M{"team_short": bson.RegEx{teamSrch, "i"}},
		}
	}
	if position != nil {
		query["pos"] = bson.RegEx{*position + "*", "i"}
	}

	players = make([]*t.Player, 0, numResults)
	sel := m.M{"score": m.M{"$meta": "textScore"}}
	err = c.Find(query).Select(sel).Sort("$textScore:score").All(&players)

	return
}
