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

func FindTweet(idStr string) (*t.Tweet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TweetsCollection())

	var tweet t.Tweet
	err := m.FindOne(c, &tweet, m.M{"id_str": idStr})
	return &tweet, err
}
