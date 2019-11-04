package db

import (
	"bet-hound/cmd/db/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertBet(bet *t.Bet) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	return m.Upsert(c, nil, m.M{"fk": bet.Fk}, m.M{"$set": bet})
}
