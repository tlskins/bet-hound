package db

import (
	"fmt"
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
	if rcp.Id != nil {
		query = m.M{"_id": *rcp.Id}
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
