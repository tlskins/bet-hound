package db

import (
	"time"

	"github.com/globalsign/mgo"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func GetGamesCurrentWeek(year int) (int, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	maxWk := []*t.GamesAggregateInt{}
	match := m.M{"$match": m.M{"yr": year}}
	group := m.M{"$group": m.M{"_id": "$yr", "value": m.M{"$max": "$wk"}}}
	pipe := []m.M{match, group}
	if err := m.Aggregate(c, &maxWk, &pipe); err != nil || len(maxWk) == 0 {
		return 0, err
	} else {
		return maxWk[0].Value, nil
	}
}

func GetMinGameResultReadyTime() (*time.Time, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	min := []*t.GamesAggregateTime{}
	match := m.M{"$match": m.M{"log": nil}}
	group := m.M{"$group": m.M{"_id": "", "value": m.M{"$min": "$gm_res_at"}}}
	pipe := []m.M{match, group}
	if err := m.Aggregate(c, &min, &pipe); err != nil || len(min) == 0 {
		return nil, err
	} else {
		return &min[0].Value, nil
	}
}

func GetResultReadyGames(leagueId string) (games []*t.GameAndLog, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	games = []*t.GameAndLog{}
	q := m.M{"lg_id": leagueId, "log": nil, "gm_res_at": m.M{"$lte": time.Now()}}
	err = m.Find(c, &games, q)
	return
}

func FindGameById(id string) (game *t.Game, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	game = &t.Game{}
	err = m.FindOne(c, game, m.M{"_id": id})
	return
}

func FindGameAndLogById(id string) (game *t.GameAndLog, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	game = &t.GameAndLog{}
	err = m.FindOne(c, game, m.M{"_id": id})
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

func UpsertGameAndLogs(games *[]*t.GameAndLog) (err error) {
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

func GamesForWeek(week, year int) (games *[]*t.Game) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	games = &[]*t.Game{}
	c.Find(m.M{"wk": week, "yr": year}).All(games)
	return
}

func SearchGames(team *string, gameTime *time.Time, week, year *int, numResults int) (games []*t.Game, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	// TODO : set indexes somewhere else
	index := mgo.Index{Key: []string{
		"$text:a_team_fk",
		"$text:a_team_name",
		"$text:h_team_fk",
		"$text:h_team_name",
	}}
	m.CreateIndex(c, index)

	query := m.M{}
	if team != nil {
		query["$text"] = m.M{"$search": *team}
	}
	if gameTime != nil {
		query["gm_time"] = *gameTime
	}
	if week != nil {
		query["wk"] = *week
	}
	if year != nil {
		query["yr"] = *year
	}

	// TODO : rewrite with pkg functions
	games = make([]*t.Game, 0, numResults)
	sel := m.M{"score": m.M{"$meta": "textScore"}}
	err = c.Find(query).Select(sel).Sort("$textScore:score").All(&games)

	return
}
