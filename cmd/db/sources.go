package db

import (
	"bet-hound/cmd/db/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertSources(sources *[]*t.Source) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.SourcesCollection())

	for _, source := range *sources {
		err = m.Upsert(c, nil, m.M{"fk": source.Fk}, m.M{"$set": source})
		if err != nil {
			return err
		}
	}
	return err
}
