package db

import (
	"github.com/globalsign/mgo/bson"
	"github.com/satori/go.uuid"

	"bet-hound/cmd/env"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
)

func UpsertUser(user *t.User) error {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())
	if user.Id == "" {
		user.Id = uuid.NewV4().String()
	}

	return m.Upsert(c, nil, m.M{"_id": user.Id}, m.M{"$set": user})
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

func FindUserById(id string) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.FindOne(c, &user, m.M{"_id": id})
	return &user, err
}

func FindUserByTwitterId(twtUserId string) (*t.User, error) {
	conn := env.MGOSession().Copy()
	defer conn.Close()
	c := conn.DB(env.MongoDb()).C(env.UsersCollection())

	var user t.User
	err := m.FindOne(c, &user, m.M{"twt.id_str": twtUserId})
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
