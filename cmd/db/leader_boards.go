package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"

	"time"
)

func UpsertLeaderBoard(leaderBoard *t.LeaderBoard) (err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.LeaderBoardsCollection())

	return m.Upsert(c, nil, m.M{"_id": leaderBoard.Id}, m.M{"$set": leaderBoard})
}

func CurrentLeaderBoards() (board *[]*t.LeaderBoard, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.LeaderBoardsCollection())

	board = &[]*t.LeaderBoard{}
	query := m.M{"st": m.M{"&gte": time.Now().AddDate(0, -1, 0)}}
	err = c.Find(query).Sort("-st").All(&board)
	return
}
