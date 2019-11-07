package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertBet(bet *t.Bet) (*t.Bet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	var result t.Bet
	err := m.Upsert(c, &result, m.M{"fk": bet.Fk}, m.M{"$set": bet})
	return &result, err
}
