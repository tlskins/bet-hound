package db

import (
	"fmt"
	"github.com/satori/go.uuid"
	"time"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func AllBets() (bets []*t.Bet) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	m.Find(c, &bets, m.M{"eqs": m.M{"$exists": true, "$ne": []m.M{}}})
	return
}

func UpsertBet(bet *t.Bet) error {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())
	if bet.Id == "" {
		bet.Id = uuid.NewV4().String()
	}

	return m.Upsert(c, nil, m.M{"_id": bet.Id}, m.M{"$set": bet})
}

func FindBetById(id string) (*t.Bet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	var bet t.Bet
	err := m.FindOne(c, &bet, m.M{"_id": id})
	return &bet, err
}

func FindBetByReply(tweet *t.Tweet) *t.Bet {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	authorId := tweet.User.IdStr
	var bet t.Bet
	fmt.Println("FindBetByReply ", tweet.InReplyToStatusIdStr, authorId)
	q := m.M{"$or": []m.M{
		m.M{"acc_fk": tweet.InReplyToStatusIdStr, "status": 0, "proposer.id_str": authorId, "pr_fk": nil},
		m.M{"acc_fk": tweet.InReplyToStatusIdStr, "status": 0, "recipient.id_str": authorId, "rr_fk": nil},
	}}
	m.FindOne(c, &bet, q)
	return &bet
}

func FindPendingFinal() []*t.Bet {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	pending := make([]*t.Bet, 0, 1)
	c.Find(m.M{"status": 1, "final_at": m.M{"$lte": time.Now()}}).All(&pending)
	return pending
}
