package db

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"
	uuid "github.com/satori/go.uuid"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertUser(user *t.User) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())
	if user.Id == "" {
		user.Id = uuid.NewV4().String()
	}

	err := m.Upsert(c, user, m.M{"_id": user.Id}, m.M{"$set": user})
	return user, err
}

func UpdateUserProfile(id string, update *t.ProfileChanges) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.Upsert(c, &user, m.M{"_id": id}, m.M{"$set": update})
	return &user, err
}

func FindUser(search string, numResults int) (users []*t.User, err error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	query := m.M{"$or": []m.M{
		m.M{"nm": bson.RegEx{search, "i"}},
		m.M{"usr_nm": bson.RegEx{search, "i"}},
		m.M{"email": bson.RegEx{search, "i"}},
	}}
	users = make([]*t.User, 0, numResults)
	err = m.Find(c, &users, query)
	return
}

func FindOrCreateBetRecipient(rcp *t.BetRecipient) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	var query m.M
	if rcp.UserId != nil {
		query = m.M{"_id": *rcp.UserId}
	} else if rcp.TwitterScreenName != nil {
		query = m.M{"twt.scrn_nm": *rcp.TwitterScreenName}
	} else {
		return nil, fmt.Errorf("No recipient provided")
	}
	if err := m.FindOne(c, &user, query); err == nil {
		return &user, nil
	}

	// create if not found and twitter name provided
	if rcp.TwitterScreenName != nil {
		newUser := t.User{
			Id: uuid.NewV4().String(),
			TwitterUser: &t.TwitterUser{
				ScreenName: *rcp.TwitterScreenName,
			},
		}
		return UpsertUser(&newUser)
	}
	return nil, fmt.Errorf("User not found")
}

func FindUserById(id string) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.FindOne(c, &user, m.M{"_id": id})
	return &user, err
}

func FindUserByIds(ids []string) ([]*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var users []*t.User
	err := m.Find(c, &users, m.M{"_id": m.M{"$in": ids}})
	return users, err
}

func SyncTwitterUser(twtUser *t.TwitterUser) error {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())
	fmt.Println("syncing... ", *twtUser)

	query := m.M{"$or": []m.M{
		m.M{"twt.scrn_nm": twtUser.ScreenName},
		m.M{"twt._id": twtUser.Id},
	}}
	var user t.User
	if err := m.FindOne(c, &user, query); err != nil {
		return err
	}

	user.TwitterUser = twtUser
	_, err := UpsertUser(&user)
	return err
}

func FindUserByTwitterId(twtUserId string) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.FindOne(c, &user, m.M{"twt.id_str": twtUserId})
	return &user, err
}

func FindUserByTwitterScreenName(screenName string) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.FindOne(c, &user, m.M{"twt.scrn_nm": screenName})
	return &user, err
}

func FindUserByUserName(userName string) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.FindOne(c, &user, m.M{"usr_nm": userName})
	return &user, err
}

func SignInUser(username, password string) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.FindOne(c, &user, m.M{"usr_nm": username, "pwd": password})
	return &user, err
}

func ViewUserProfile(userId string) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.Upsert(c, &user, m.M{"_id": userId}, m.M{"$set": m.M{"lst_vw": time.Now()}})
	return &user, err
}

func SyncBetWithUsers(event string, bet *t.Bet) (*t.Notification, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	// notification
	var note *t.Notification
	var pUpdate, rUpdate m.M

	if event == "Create" {
		note, pUpdate, rUpdate = buildBetCreatedNote(bet)
	} else if event == "Update" {
		note, pUpdate, rUpdate = buildBetUpdatedNote(bet)
	} else if event == "Final" {
		note, pUpdate, rUpdate = buildBetFinalNote(bet)
	}

	if err := c.Update(m.M{"_id": bet.Proposer.Id}, pUpdate); err != nil {
		return nil, err
	}
	if bet.Recipient != nil {
		if err := c.Update(m.M{"_id": bet.Recipient.Id}, rUpdate); err != nil {
			return nil, err
		}
	}

	return note, nil
}

func buildBetCreatedNote(bet *t.Bet) (note *t.Notification, pUpdate, rUpdate m.M) {
	note = &t.Notification{
		Id:      uuid.NewV4().String(),
		SentAt:  time.Now(),
		Title:   fmt.Sprintf("%s proposed a bet with %s", bet.ProposerName(), bet.RecipientName()),
		Type:    "BetCreated",
		Message: bet.String(),
	}
	pUpdate = m.M{"$push": m.M{
		"pnd_t_bts": bet.Id,
		"notes":     m.M{"$each": []t.Notification{*note}, "$slice": 10, "$sort": m.M{"snt_at": -1}},
	}}
	rUpdate = m.M{"$push": m.M{
		"pnd_u_bts": bet.Id,
		"notes":     m.M{"$each": []t.Notification{*note}, "$slice": 10, "$sort": m.M{"snt_at": -1}},
	}}
	return
}

func buildBetUpdatedNote(bet *t.Bet) (note *t.Notification, pUpdate, rUpdate m.M) {
	note = &t.Notification{
		Id:      uuid.NewV4().String(),
		SentAt:  time.Now(),
		Title:   fmt.Sprintf("%s's bet with %s was %s", bet.ProposerName(), bet.RecipientName(), bet.BetStatus.String()),
		Type:    "BetUpdated",
		Message: bet.BetStatus.String() + ": " + bet.String(),
	}

	var prgBetId *string
	if bet.BetStatus.String() == "Accepted" {
		prgBetId = &bet.Id
	}
	pUpdate = m.M{
		"$push": m.M{
			"notes":   m.M{"$each": []t.Notification{*note}, "$slice": 10, "$sort": m.M{"snt_at": -1}},
			"prg_bts": prgBetId,
		},
		"$pull": m.M{"pnd_u_bts": bet.Id, "pnd_t_bts": bet.Id},
	}
	rUpdate = m.M{
		"$push": m.M{
			"notes":   m.M{"$each": []t.Notification{*note}, "$slice": 10, "$sort": m.M{"snt_at": -1}},
			"prg_bts": prgBetId,
		},
		"$pull": m.M{"pnd_u_bts": bet.Id, "pnd_t_bts": bet.Id},
	}
	return
}

func buildBetFinalNote(bet *t.Bet) (note *t.Notification, pUpdate, rUpdate m.M) {
	note = &t.Notification{
		Id:      uuid.NewV4().String(),
		SentAt:  time.Now(),
		Title:   fmt.Sprintf("%s's bet with %s is %s", bet.ProposerName(), bet.RecipientName(), bet.BetStatus.String()),
		Type:    "BetFinal",
		Message: bet.ResultString(),
	}

	pUpdate = m.M{
		"$push": m.M{"notes": m.M{"$each": []t.Notification{*note}, "$slice": 10, "$sort": m.M{"snt_at": -1}}},
		"$pull": m.M{"prg_bts": bet.Id},
	}
	rUpdate = m.M{
		"$push": m.M{"notes": m.M{"$each": []t.Notification{*note}, "$slice": 10, "$sort": m.M{"snt_at": -1}}},
		"$pull": m.M{"prg_bts": bet.Id},
	}

	if bet.BetResult.Winner.Id == bet.Proposer.Id {
		pUpdate["$inc"] = m.M{"bts_wn": 1}
		rUpdate["$inc"] = m.M{"bts_lst": 1}
	} else {
		rUpdate["$inc"] = m.M{"bts_wn": 1}
		pUpdate["$inc"] = m.M{"bts_lst": 1}

	}
	return
}
