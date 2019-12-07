package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertTweet(tweet *t.Tweet) error {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TweetsCollection())

	return m.Upsert(c, nil, m.M{"_id": tweet.Id}, m.M{"$set": tweet})
}

func FindTweet(id int64) (*t.Tweet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TweetsCollection())

	var tweet t.Tweet
	err := m.FindOne(c, &tweet, m.M{"_id": id})
	return &tweet, err
}
