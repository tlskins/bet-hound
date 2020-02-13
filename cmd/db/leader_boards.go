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

func CurrentLeaderBoards() ([]*t.LeaderBoard, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.LeaderBoardsCollection())

	match := m.M{"st": m.M{"$gte": time.Now().AddDate(0, -1, 0)}}
	unwind := "$ldrs"
	lookup := m.M{
		"from": env.UsersCollection(),
		"let":  m.M{"ldr_id": "$ldrs.usr_id"},
		"pipeline": []m.M{m.M{
			"$match": m.M{"$expr": m.M{"$eq": []interface{}{"$_id", "$$ldr_id"}}},
		}},
		"as": "ldrs.usr",
	}
	unwindLk := "$ldrs.usr"
	group := m.M{
		"_id":   "$_id",
		"lg_id": m.M{"$first": "$lg_id"},
		"st":    m.M{"$first": "$st"},
		"end":   m.M{"$first": "$end"},
		"fin":   m.M{"$first": "$fin"},
		"ldrs":  m.M{"$push": "$ldrs"},
	}
	sort := m.M{"st": -1}

	boards := []*t.LeaderBoard{}
	if err := m.Aggregate(c, &boards, []m.M{
		m.M{"$match": match},
		m.M{"$unwind": unwind},
		m.M{"$lookup": lookup},
		m.M{"$unwind": unwindLk},
		m.M{"$group": group},
		m.M{"$sort": sort},
	}); err != nil {
		return nil, err
	}
	return boards, nil
}
