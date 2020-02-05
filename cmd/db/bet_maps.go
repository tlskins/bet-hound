package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"

	"github.com/globalsign/mgo/bson"
)

func UpsertBetMaps(betMaps *[]*t.BetMap) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetMapsCollection())

	for _, betMap := range *betMaps {
		if err = m.Upsert(c, nil, m.M{"_id": betMap.Id}, m.M{"$set": betMap}); err != nil {
			return err
		}
	}
	return
}

func SearchBetMaps(betType, search string, numResults int) (users []*t.User, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	users = make([]*t.User, 0, numResults)
	err = m.Find(c, &users, m.M{"t": betType, "n": bson.RegEx{search, "i"}})
	return
}
