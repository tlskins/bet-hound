package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertTweet(tweet *t.Tweet) (*t.Tweet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TweetsCollection())

	var result t.Tweet
	err := m.Upsert(c, &result, m.M{"_id": tweet.Id}, m.M{"$set": tweet})
	return &result, err
}

func FindTweet(id int64) (*t.Tweet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TweetsCollection())

	var tweet t.Tweet
	err := m.FindOne(c, &tweet, m.M{"_id": id})
	return &tweet, err
}
