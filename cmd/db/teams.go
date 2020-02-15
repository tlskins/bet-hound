package db

import (
	"time"

	"github.com/globalsign/mgo/bson"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertTeams(teams *[]*t.Team) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TeamsCollection())

	for _, team := range *teams {
		err = m.Upsert(c, nil, m.M{"_id": team.Id}, m.M{"$set": team})
		if err != nil {
			return err
		}
	}
	return err
}

func FindTeamById(id string) (team *t.Team, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TeamsCollection())

	team = &t.Team{}
	err = m.FindOne(c, team, m.M{"_id": id})
	return
}

func SearchTeamsWithGame(name, location *string, numResults int) (teams []*t.Team, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TeamsCollection())

	// match pipe
	matches := []m.M{}
	if name != nil {
		matches = append(matches, m.M{"nm": bson.RegEx{*name, "i"}})
	}
	if location != nil {
		matches = append(matches, m.M{"loc": bson.RegEx{*location, "i"}})
	}
	match := m.M{"$or": matches}

	// lookup pipe
	lookup := m.M{
		"from": env.GamesCollection(),
		"let":  m.M{"tm_fk": "$fk", "p_lg_id": "$lg_id"},
		"pipeline": []m.M{
			m.M{"$match": m.M{"$expr": m.M{"$and": []m.M{
				m.M{"$eq": []interface{}{"$lg_id", "$$p_lg_id"}},
				m.M{"$gt": []interface{}{"$gm_time", time.Now()}},
				m.M{"$or": []m.M{
					m.M{"$eq": []interface{}{"$a_team_fk", "$$tm_fk"}},
					m.M{"$eq": []interface{}{"$h_team_fk", "$$tm_fk"}},
				}},
			}}}},
			m.M{"$sort": m.M{"gm_time": 1}},
			m.M{"$limit": 1},
		},
		"as": "lk_gms",
	}

	// addfield pipe
	addField := m.M{"gm": m.M{"$arrayElemAt": []interface{}{"$lk_gms", 0}}}

	// filter no games
	match2 := m.M{"gm": m.M{"$ne": nil}}

	teams = make([]*t.Team, 0, numResults)
	err = m.Aggregate(c, &teams, []m.M{
		m.M{"$match": match},
		m.M{"$lookup": lookup},
		m.M{"$addFields": addField},
		m.M{"$match": match2},
		m.M{"$limit": numResults},
	})
	return
}
