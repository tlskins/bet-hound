package db

import (
	"fmt"
	"github.com/satori/go.uuid"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

// func CreateBet(bet *t.Bet) (*t.Bet, error) {
// 	conn := env.MGOSession().Copy()
// 	defer conn.Close()
// 	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

// 	err := m.Upsert(c, &bet, nil, m.M{"$set": bet})
// 	return bet, err
// }

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

func FindBetByReply(tweet *t.Tweet) (*t.Bet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	authorId := tweet.User.IdStr
	var bet t.Bet
	fmt.Println("FindBetByReply ", tweet.IdStr, authorId)
	q := m.M{"$or": []m.M{
		m.M{"pchk_tweet_id": tweet.IdStr, "status": 0, "proposer.id_str": authorId},
		m.M{"rchk_tweet_id": tweet.IdStr, "status": 1, "recipient.id_str": authorId},
	}}
	// q := m.M{"$and": [ []m.M{"$or": [m.M{"pchk_tweet_id": tweetId}, m.M{"rchk_tweet_id": tweetId}]}, m.M{"$or": [m.M{"status": 0}, m.M{"status": 1}]} ]}
	err := m.FindOne(c, &bet, q)
	return &bet, err
}
