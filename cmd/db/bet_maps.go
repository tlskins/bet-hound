package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"

	"github.com/globalsign/mgo/bson"
)

func UpsertBetMaps(betMaps *[]*t.BetMap) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetMapsCollection())

	for _, betMap := range *betMaps {
		if err = m.Upsert(c, nil, m.M{"_id": betMap.Id}, m.M{"$set": betMap}); err != nil {
			return err
		}
	}
	return
}

func SearchBetMaps(betType, search string) (betMaps []*t.BetMap, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetMapsCollection())

	betMaps = []*t.BetMap{}
	err = m.Find(c, &betMaps, m.M{"t": betType, "n": bson.RegEx{search, "i"}})
	return
}

func GetBetMaps(leagueId, betType *string) (betMaps []*t.BetMap, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetMapsCollection())

	q := m.M{}
	if leagueId != nil {
		q["lg_id"] = m.M{"$in": []string{*leagueId, "*"}}
	}
	if betType != nil {
		q["t"] = m.M{"$in": []string{*betType, "*"}}
	}

	betMaps = []*t.BetMap{}
	err = m.Find(c, &betMaps, q)
	return
}
