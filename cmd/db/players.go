package db

import (
	"time"

	"github.com/globalsign/mgo/bson"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertPlayers(players *[]*t.Player) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	for _, player := range *players {
		err = m.Upsert(c, nil, m.M{"_id": player.Id}, m.M{"$set": player})
		if err != nil {
			return err
		}
	}
	return err
}

func FindPlayerById(id string) (player *t.Player, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	player = &t.Player{}
	err = m.FindOne(c, player, m.M{"_id": id})
	return
}

func FindTeamRoster(leagueId, teamFk string) (players []*t.Player, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	players = []*t.Player{}
	err = m.Find(c, &players, m.M{"team_fk": teamFk, "lg_id": leagueId})
	return
}

func SearchPlayersWithGame(name, team, position *string, numResults int) (players []*t.Player, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	// match pipe
	match := m.M{}
	if name != nil {
		match["name"] = bson.RegEx{*name, "i"}
	}
	if team != nil {
		teamSrch := *team + "*"
		match["$or"] = []m.M{
			m.M{"team_fk": bson.RegEx{teamSrch, "i"}},
			m.M{"team_name": bson.RegEx{teamSrch, "i"}},
			m.M{"team_short": bson.RegEx{teamSrch, "i"}},
		}
	}
	if position != nil {
		match["pos"] = bson.RegEx{*position + "*", "i"}
	}

	// lookup pipe
	lookup := m.M{
		"from": env.GamesCollection(),
		"let":  m.M{"tm_fk": "$team_fk", "p_lg_id": "$lg_id"},
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

	players = make([]*t.Player, numResults)
	err = m.Aggregate(c, &players, []m.M{
		m.M{"$match": match},
		m.M{"$lookup": lookup},
		m.M{"$addFields": addField},
		m.M{"$match": match2},
		m.M{"$limit": numResults},
	})
	return
}
