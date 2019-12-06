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
		err = m.Upsert(c, player, m.M{"_id": player.Id}, m.M{"$set": player})
		// err = m.Upsert(c, nil, nil, m.M{"$set": player})
		if err != nil {
			return err
		}
	}
	return err
}

func SearchPlayerByName(search string) (player *t.Player) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	// TODO : set indexes somewhere else
	index := mgo.Index{Key: []string{"$text:f_name", "$text:l_name"}}
	m.CreateIndex(c, index)

	// TODO : rewrite with pkg functions
	result := make([]t.Player, 0, 1)
	query := m.M{"$text": m.M{"$search": search}}
	sel := m.M{"score": m.M{"$meta": "textScore"}}
	q := c.Find(query).Select(sel).Sort("$textScore:score")
	q.All(&result)

	if len(result) > 0 {
		return &result[0]
	} else {
		return nil
	}
}
