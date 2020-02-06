package db

import (
	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"

	uuid "github.com/satori/go.uuid"
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

	mBets := []*t.MongoBet{}
	if err = c.Find(q).Sort("-crt_at").All(&mBets); err != nil {
		return
	}

	return convertMongoBets(mBets)
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
	}}

	mBets := []*t.MongoBet{}
	if err = c.Find(q).Sort("-crt_at").All(&mBets); err != nil {
		return
	}

	return convertMongoBets(mBets)
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

// func FindAcceptedBetsByGame(gameId string) ([]*t.Bet, error) {
// 	conn := env.MGOSession().Copy()
// 	defer conn.Close()
// 	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

// 	mBets := []*t.MongoBet{}
// 	q := m.M{
// 		"status":    1,
// 		"eqs.exprs": m.M{"$elemMatch": m.M{"gm._id": gameId}},
// 	}
// 	if err := m.Find(c, &mBets, q); err != nil {
// 		return nil, err
// 	}
// 	return convertMongoBets(mBets)
// }

func FindBetById(id string) (*t.Bet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	var mBet t.MongoBet
	err := m.FindOne(c, &mBet, m.M{"_id": id})
	return mBet.Bet(), err
}

func FindBetByReply(tweet *t.Tweet) (*t.Bet, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	var mBet t.MongoBet
	err := m.FindOne(c, &mBet, m.M{"acc_fk": tweet.InReplyToStatusIdStr, "status": 0})
	return mBet.Bet(), err
}

// helpers

func convertMongoBets(mBets []*t.MongoBet) (bets []*t.Bet, err error) {
	bets = make([]*t.Bet, len(mBets))
	for i, mBet := range mBets {
		bets[i] = mBet.Bet()
	}
	return bets, nil
}
