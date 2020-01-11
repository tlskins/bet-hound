package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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

func FindPlayer(fk string) (*t.Player, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	var player t.Player
	err := m.FindOne(c, player, m.M{"fk": fk})
	return &player, err
}

func SearchPlayers(name, team, position *string, numResults int) (players []*t.Player, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	// TODO : set indexes somewhere else
	index := mgo.Index{Key: []string{"$text:f_name", "$text:l_name", "$text:name"}}
	m.CreateIndex(c, index)

	query := m.M{}
	if name != nil {
		// query["$text"] = m.M{"$search": *name}
		query["name"] = bson.RegEx{*name, "i"}
	}
	if team != nil {
		teamSrch := *team + "*"
		query["$or"] = []m.M{
			m.M{"team_fk": bson.RegEx{teamSrch, "i"}},
			m.M{"team_name": bson.RegEx{teamSrch, "i"}},
			m.M{"team_short": bson.RegEx{teamSrch, "i"}},
		}
	}
	if position != nil {
		query["pos"] = bson.RegEx{*position + "*", "i"}
	}

	players = make([]*t.Player, 0, numResults)
	sel := m.M{"score": m.M{"$meta": "textScore"}}
	err = c.Find(query).Select(sel).Sort("$textScore:score").All(&players)

	return
}

func SearchPlayersWithGame(name, team, position *string, numResults int) (players []*t.Player, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.PlayersCollection())

	settings, err := GetLeagueSettings("nfl")
	if err != nil {
		return players, err
	}

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
		"let":  m.M{"tm_fk": "$team_fk"},
		"pipeline": []m.M{
			m.M{"$match": m.M{"$expr": m.M{"$and": []m.M{
				m.M{"$eq": []interface{}{"$yr", settings.CurrentYear}},
				m.M{"$eq": []interface{}{"$wk", settings.CurrentWeek}},
				m.M{"$or": []m.M{
					m.M{"$eq": []interface{}{"$a_team_fk", "$$tm_fk"}},
					m.M{"$eq": []interface{}{"$h_team_fk", "$$tm_fk"}},
				}},
			}}}},
		},
		"as": "lk_gms",
	}

	// addfield pipe
	addField := m.M{"gm": m.M{"$arrayElemAt": []interface{}{"$lk_gms", 0}}}

	players = make([]*t.Player, 0, numResults)
	err = m.Aggregate(c, &players, []m.M{
		m.M{"$match": match},
		m.M{"$lookup": lookup},
		m.M{"$addFields": addField},
	})
	return
}
