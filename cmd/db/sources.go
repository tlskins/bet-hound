package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
	"fmt"
	"github.com/globalsign/mgo"
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
	fmt.Println("SearchSourceByName", env.MongoDb(), env.SourcesCollection(), env.MGOSession())
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.SourcesCollection())

	// TODO : set indexes somewhere else
	index := mgo.Index{Key: []string{"$text:f_name", "$text:l_name"}}
	m.CreateIndex(c, index)

	// TODO : rewrite with pkg functions
	result = make([]t.Source, 0, numResults)
	query := m.M{"$text": m.M{"$search": search}}
	sel := m.M{"score": m.M{"$meta": "textScore"}}
	q := c.Find(query).Select(sel).Sort("$textScore:score")
	err = q.All(&result)
	return result, err
}
