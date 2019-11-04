package db

import (
	"bet-hound/cmd/db/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
	"regexp"
	// "github.com/globalsign/mgo"
	// "github.com/globalsign/mgo/bson"
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

func SearchSourceByName(search string, numResults int) (result []t.Source, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.SourcesCollection())

	// index := mgo.Index{Key: []string{"$text:name"}}
	// m.CreateIndex(c, index)

	result = make([]t.Source, 0, numResults)
	// err = m.Find(c, &result, m.M{"$text": m.M{"$search": search}})
	var r = regexp.MustCompile("(?i)" + search)
	err = m.Find(c, &result, m.M{"name": m.M{"$regex": r.String()}})
	return result, err
}
