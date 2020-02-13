package db

import (
	"bet-hound/cmd/env"
	"bet-hound/cmd/types"
	t "bet-hound/cmd/types"
	"bet-hound/pkg/helpers"
	m "bet-hound/pkg/mongo"
	"fmt"

	"time"

	"github.com/globalsign/mgo/bson"
	uuid "github.com/satori/go.uuid"
)

func Bets(userId string) (resp *t.BetsResponse, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())
	q := m.M{"$or": []m.M{
		m.M{"proposer._id": userId},
		m.M{"recipient._id": userId},
	}}

	mBets := []*t.MongoBet{}
	if err = c.Find(q).Sort("-crt_at").All(&mBets); err != nil {
		return
	}
	if bets, convErr := convertMongoBets(mBets); err != nil {
		return resp, convErr
	} else {
		resp = groupBetsResponse(&bets)
	}
	return
}

func CurrentBets() (resp *t.BetsResponse, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())
	q := m.M{"$or": []m.M{
		m.M{"$and": []m.M{
			m.M{"status": 0},
			m.M{"recipient": nil},
		}},
		m.M{"$or": []m.M{
			m.M{"status": 1},
			m.M{"status": 2},
		}},
	}}

	mBets := []*t.MongoBet{}
	if err = c.Find(q).Sort("-crt_at").All(&mBets); err != nil {
		return
	}
	if bets, convErr := convertMongoBets(mBets); err != nil {
		return resp, convErr
	} else {
		resp = groupBetsResponse(&bets)
	}
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

func SearchBets(search string, userId, betStatus *string) (bets []*t.Bet, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	searchQueries := []m.M{
		m.M{"proposer.nm": bson.RegEx{search, "i"}},
		m.M{"proposer.usr_nm": bson.RegEx{search, "i"}},
		m.M{"proposer.twt.nm": bson.RegEx{search, "i"}},
		m.M{"proposer.twt.scrn_nm": bson.RegEx{search, "i"}},
		m.M{"recipient.nm": bson.RegEx{search, "i"}},
		m.M{"recipient.usr_nm": bson.RegEx{search, "i"}},
		m.M{"recipient.twt.nm": bson.RegEx{search, "i"}},
		m.M{"recipient.twt.scrn_nm": bson.RegEx{search, "i"}},
		m.M{"recipient.twt.scrn_nm": bson.RegEx{search, "i"}},
		m.M{"eqs.exprs.tm.nm": bson.RegEx{search, "i"}},
		m.M{"eqs.exprs.tm.loc": bson.RegEx{search, "i"}},
		m.M{"eqs.exprs.tm.fk": bson.RegEx{search, "i"}},
		m.M{"eqs.exprs.player.name": bson.RegEx{search, "i"}},
		m.M{"eqs.exprs.player.team_name": bson.RegEx{search, "i"}},
		m.M{"eqs.exprs.player.team_fk": bson.RegEx{search, "i"}},
		m.M{"eqs.exprs.player.pos": bson.RegEx{search, "i"}},
	}

	var query m.M
	if userId != nil || betStatus != nil {
		and := []m.M{m.M{"$or": searchQueries}}
		if userId != nil {
			and = append(and, m.M{"$or": []m.M{
				m.M{"proposer._id": *userId},
				m.M{"recipient._id": *userId},
			}})
		}
		if betStatus != nil {
			code := t.BetStatusFromString(*betStatus)
			and = append(and, m.M{"status": int(code)})
		}
		query = m.M{"$and": and}
	} else {
		query = m.M{"$or": searchQueries}
	}

	mBets := []*t.MongoBet{}
	if err = c.Find(query).Sort("-crt_at").All(&mBets); err != nil {
		return
	}
	return convertMongoBets(mBets)
}

func GetResultReadyBets(leagueId string) (bets []*t.Bet, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	q := m.M{
		"lg_id":    leagueId,
		"final_at": m.M{"$lte": time.Now()},
		"rslt":     nil,
		"status":   1,
	}

	mBets := []*t.MongoBet{}
	m.Find(c, &mBets, q)

	return convertMongoBets(mBets)
}

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

func BuildLeaderBoard(start, end *time.Time, leagueId string) (*t.LeaderBoard, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.BetsCollection())

	match := m.M{
		"lg_id":    leagueId,
		"final_at": m.M{"$gte": start, "$lt": end},
		"status":   t.BetStatusFinal,
		"rslt":     m.M{"$ne": nil},
	}
	addField := m.M{"user_id": []string{"$proposer._id", "$recipient._id"}}
	unwind := "$user_id"
	group := m.M{
		"_id": "$user_id",
		"scr": m.M{"$sum": m.M{
			"$cond": []interface{}{
				m.M{"$eq": []string{"$user_id", "$rslt.winner._id"}},
				1,
				-0.5,
			},
		}},
		"ws": m.M{"$sum": m.M{
			"$cond": []interface{}{
				m.M{"$eq": []string{"$user_id", "$rslt.winner._id"}},
				1,
				0,
			},
		}},
		"ls": m.M{"$sum": m.M{
			"$cond": []interface{}{
				m.M{"$eq": []string{"$user_id", "$rslt.loser._id"}},
				1,
				0,
			},
		}},
		"w_bets": m.M{"$addToSet": m.M{
			"$cond": []interface{}{
				m.M{"$eq": []string{"$user_id", "$rslt.winner._id"}},
				"$_id",
				nil,
			},
		}},
		"l_bets": m.M{"$addToSet": m.M{
			"$cond": []interface{}{
				m.M{"$eq": []string{"$user_id", "$rslt.loser._id"}},
				"$_id",
				nil,
			},
		}},
	}
	sort := m.M{"scr": -1}
	project := m.M{
		"_id":    0,
		"usr_id": "$_id",
		"scr":    1,
		"ws":     1,
		"ls":     1,
		"w_bets": m.M{
			"$filter": m.M{
				"input": "$w_bets",
				"as":    "w_bet",
				"cond":  m.M{"$ne": []interface{}{"$$w_bet", nil}},
			},
		},
		"l_bets": m.M{
			"$filter": m.M{
				"input": "$l_bets",
				"as":    "l_bet",
				"cond":  m.M{"$ne": []interface{}{"$$l_bet", nil}},
			},
		},
	}

	var leaders []*t.Leader
	if err := m.Aggregate(c, &leaders, []m.M{
		m.M{"$match": match},
		m.M{"$addFields": addField},
		m.M{"$unwind": unwind},
		m.M{"$group": group},
		m.M{"$sort": sort},
		m.M{"$project": project},
	}); err != nil {
		return nil, err
	}

	for i, leader := range leaders {
		leader.Rank = i + 1
	}

	leaderBoard := t.LeaderBoard{
		Id:        genLeaderBoardId(leagueId, start, end),
		LeagueId:  leagueId,
		StartTime: *start,
		EndTime:   *end,
		Leaders:   leaders,
	}

	fmt.Println(helpers.PrettyPrint(leaderBoard))
	return &leaderBoard, nil
}

// helpers

func genLeaderBoardId(leagueId string, start, end *time.Time) string {
	return fmt.Sprintf(
		"%s%s%s",
		leagueId,
		start.Format("20060102"),
		end.Format("20060102"),
	)
}

func groupBetsResponse(bets *[]*t.Bet) *t.BetsResponse {
	acceptedBets := []*types.Bet{}
	finalBets := []*types.Bet{}
	pendingBets := []*types.Bet{}
	publicPendingBets := []*types.Bet{}
	closedBets := []*types.Bet{}

	for _, bet := range *bets {
		if bet.BetStatus.String() == "Final" {
			finalBets = append(finalBets, bet)
		} else if bet.BetStatus.String() == "Accepted" {
			acceptedBets = append(acceptedBets, bet)
		} else if bet.BetStatus.String() == "Pending Approval" && bet.Recipient == nil {
			publicPendingBets = append(publicPendingBets, bet)
		} else if bet.BetStatus.String() == "Pending Approval" && bet.Recipient != nil {
			pendingBets = append(pendingBets, bet)
		} else {
			closedBets = append(closedBets, bet)
		}
	}

	return &t.BetsResponse{
		AcceptedBets:      acceptedBets,
		FinalBets:         finalBets,
		PendingBets:       pendingBets,
		PublicPendingBets: publicPendingBets,
		ClosedBets:        closedBets,
	}
}

func convertMongoBets(mBets []*t.MongoBet) (bets []*t.Bet, err error) {
	bets = make([]*t.Bet, len(mBets))
	for i, mBet := range mBets {
		bets[i] = mBet.Bet()
	}
	return bets, nil
}
