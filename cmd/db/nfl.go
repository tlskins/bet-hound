package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
	"fmt"
)

func UpsertLeagueSettings(settings *t.LeagueSettings) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.LeagueSettingsCollection())

	return m.Upsert(c, settings, m.M{"_id": settings.Id}, m.M{"$set": settings})
}

func GetLeagueSettings(id string) (settings *t.LeagueSettings, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.LeagueSettingsCollection())

	matches := make([]*t.LeagueSettings, 1, 1)
	if err := c.Find(m.M{"_id": id}).All(&matches); err != nil {
		return nil, err
	}

	if len(matches) > 0 {
		return matches[0], nil
	} else {
		return nil, fmt.Errorf("No settings found")
	}
}
