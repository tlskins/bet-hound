package db

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func FindGameById(id string) (game *t.Game, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	game = &t.Game{}
	err = m.FindOne(c, game, m.M{"_id": id})
	return
}

func FindGameAndLogById(id, leagueId string) (game *t.GameAndLog, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	if leagueId == "nba" {
		nbaGame := t.NbaGameAndLog{}
		if err = m.FindOne(c, &nbaGame, m.M{"_id": id}); err != nil {
			return nil, err
		}
		var gmAndLog t.GameAndLog = nbaGame
		game = &gmAndLog
	} else {
		return nil, fmt.Errorf("Unable to find game and log by id for leagueId: %s", leagueId)
	}
	return
}

func GetCurrentGames() (games []*t.Game, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.TeamsCollection())

	lookup := m.M{
		"from": env.GamesCollection(),
		"let":  m.M{"tm_fk": "$fk"},
		"pipeline": []m.M{
			m.M{"$match": m.M{"$expr": m.M{"$and": []m.M{
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

	games = []*t.Game{}
	err = m.Aggregate(c, &games, []m.M{
		m.M{"$lookup": lookup},
		m.M{"$unwind": "$lk_gms"},
		m.M{"$replaceRoot": m.M{"newRoot": "$lk_gms"}},
		m.M{"$group": m.M{"_id": "$_id", "data": m.M{"$addToSet": "$$ROOT"}}},
		m.M{"$replaceRoot": m.M{"newRoot": m.M{"$arrayElemAt": []interface{}{"$data", 0}}}},
		m.M{"$sort": m.M{"gm_time": 1}},
	})
	return
}

func UpsertGames(games *[]*t.Game) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	for _, game := range *games {
		err = m.Upsert(c, game, m.M{"_id": game.Id}, m.M{"$set": game})
		if err != nil {
			return err
		}
	}
	return err
}

func UpsertGameLog(gameId string, gameLog *t.GameLog) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	return m.Upsert(c, nil, m.M{"_id": gameId}, m.M{"$set": m.M{"log": gameLog}})
}

func GetResultReadyGames(leagueId string) (games []*t.Game, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	games = []*t.Game{}
	q := m.M{"lg_id": leagueId, "log": nil, "gm_res_at": m.M{"$lte": time.Now()}}
	err = m.Find(c, &games, q)
	return
}

func SearchGames(team *string, gameTime *time.Time, numResults int) (games []*t.Game, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	andQuery := []m.M{}
	if team != nil {
		andQuery = append(andQuery, m.M{"$or": []m.M{
			m.M{"a_team_name": bson.RegEx{*team, "i"}},
			m.M{"h_team_name": bson.RegEx{*team, "i"}},
		}})
	}
	if gameTime != nil {
		minTime := time.Date(gameTime.Year(), gameTime.Month(), gameTime.Day(), 0, 0, 0, 0, env.TimeZone())
		maxTime := minTime.Add(24 * time.Hour)
		andQuery = append(andQuery, m.M{"gm_time": m.M{"$gte": minTime, "$lte": maxTime}})
	}
	query := m.M{"$and": andQuery}

	games = make([]*t.Game, 0, numResults)
	err = m.Find(c, &games, query)
	return
}
