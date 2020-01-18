package db

import (
	"time"

	"github.com/globalsign/mgo"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertCurrentGames(games *[]*t.Game) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.CurrentGamesCollection())

	for _, game := range *games {
		err = m.Upsert(c, game, m.M{"_id": game.Id}, m.M{"$set": game})
		if err != nil {
			return err
		}
	}
	return err
}

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

func GetResultReadyGames() (games []*t.Game, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	games = []*t.Game{}
	err = m.Find(c, &games, m.M{"log": nil, "gm_res_at": m.M{"$lte": time.Now()}})
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

func FindCurrentGame(settings *t.LeagueSettings, fk string) (game *t.Game, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.GamesCollection())

	game = &t.Game{}
	err = m.FindOne(c, game, m.M{"fk": fk, "wk": settings.CurrentWeek, "yr": settings.CurrentYear})
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
