package db

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func Bets(userId string) (bets []*t.Bet, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())
	q := m.M{"$and": []m.M{
		m.M{"$or": []m.M{
			m.M{"proposer._id": userId},
			m.M{"recipient._id": userId},
		}},
		m.M{"eqs": m.M{"$exists": true, "$ne": []m.M{}}},
	}}

	bets = []*t.Bet{}
	err = c.Find(q).Sort("-crt_at").All(&bets)
	return
}

func CurrentBets() (bets []*t.Bet, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())
	q := m.M{"$and": []m.M{
		m.M{"$or": []m.M{
			m.M{"status": 1},
			m.M{"status": 2},
		}},
		m.M{"eqs": m.M{"$exists": true, "$ne": []m.M{}}},
	}}

	bets = []*t.Bet{}
	err = c.Find(q).Sort("-crt_at").All(&bets)
	return
}

func UpsertBet(bet *t.Bet) error {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())
	if bet.Id == "" {
		bet.Id = uuid.NewV4().String()
	}

	return m.Upsert(c, nil, m.M{"_id": bet.Id}, m.M{"$set": bet})
}

func FindAcceptedBetsByGame(gameId string) (*[]*t.Bet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	bets := []*t.Bet{}
	q := m.M{
		"status": 1,
		"eqs": m.M{"$elemMatch": m.M{
			"exprs": m.M{
				"$elemMatch": m.M{
					"gm._id": gameId,
				},
			},
		}},
	}
	err := m.Find(c, &bets, q)
	return &bets, err
}

func FindBetById(id string) (*t.Bet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	var bet t.Bet
	err := m.FindOne(c, &bet, m.M{"_id": id})
	return &bet, err
}

func FindBetByReply(tweet *t.Tweet) (*t.Bet, error) {
	if tweet.InReplyToStatusIdStr == "" {
		return nil, fmt.Errorf("Tweet doesnt reply to a bet")
	} else if tweet.TwitterUser.IdStr == "" {
		return nil, fmt.Errorf("Tweet doest not have an author")
	}

	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	authorId := tweet.TwitterUser.IdStr
	var bet t.Bet
	q := m.M{"$or": []m.M{
		m.M{"acc_fk": tweet.InReplyToStatusIdStr, "status": 0, "proposer.twt.id_str": authorId, "pr_fk": nil},
		m.M{"acc_fk": tweet.InReplyToStatusIdStr, "status": 0, "recipient.twt.id_str": authorId, "rr_fk": nil},
	}}
	err := m.FindOne(c, &bet, q)
	return &bet, err
}

func FindPendingFinalBets() []*t.Bet {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	pending := make([]*t.Bet, 0, 1)
	c.Find(m.M{"status": 1, "final_at": m.M{"$lte": time.Now()}}).All(&pending)
	return pending
}
