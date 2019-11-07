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
	err := m.Upsert(c, &result, m.M{"_id": bet.Id}, m.M{"$set": bet})
	return &result, err
}

func FindBetByProposerCheckTweet(tweetId string) (*t.Bet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TweetsCollection())

	var bet t.Bet
	err := m.FindOne(c, &bet, m.M{"pchk_tweet_id": tweetId, "status": 0})
	return &bet, err
}
