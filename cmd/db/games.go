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

func GetCurrentGames() (games *[]*t.Game) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.CurrentGamesCollection())

	games = &[]*t.Game{}
	c.Find(m.M{}).All(games)
	return
}

// func GetCurrentWeek() (yr, wk int) {
// 	conn := env.MGOSession().Copy()
// 	defer conn.Close()
// 	c := conn.DB(env.MongoDb()).C(env.CurrentGamesCollection())

// 	games := make([]*t.Game, 1)
// 	c.Find(m.M{}).Limit(1).All(&games)
// 	if len(games) == 1 {
// 		return games[0].Year, games[0].Week
// 	} else {
// 		return 0, 0
// 	}
// 	return
// }

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
