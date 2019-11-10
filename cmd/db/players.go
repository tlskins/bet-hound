package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
	"github.com/globalsign/mgo"
	"github.com/satori/go.uuid"
)

func UpsertPlayers(players *[]*t.Player) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	for _, player := range *players {
		if player.Id == "" {
			player.Id = uuid.NewV4().String()
		}
		err = m.Upsert(c, nil, m.M{"fk": player.Fk}, m.M{"$set": player})
		if err != nil {
			return err
		}
	}
	return err
}

func SearchPlayerByName(search string, numResults int) (result []t.Player) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	// TODO : set indexes somewhere else
	index := mgo.Index{Key: []string{"$text:f_name", "$text:l_name"}}
	m.CreateIndex(c, index)

	// TODO : rewrite with pkg functions
	result = make([]t.Player, 0, numResults)
	query := m.M{"$text": m.M{"$search": search}}
	sel := m.M{"score": m.M{"$meta": "textScore"}}
	q := c.Find(query).Select(sel).Sort("$textScore:score")
	_ = q.All(&result)
	return result
}
