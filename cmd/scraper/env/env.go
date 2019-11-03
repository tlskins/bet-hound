package env

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/spf13/viper"

	"bet-hound/pkg/mongo"
)

type environment struct {
	m *mgo.Session

	port              string
	mongoHost         string
	mongoUser         string
	mongoPwd          string
	mongoDb           string
	sourcesCollection string
	logPath           string
}

var e = &environment{}

func Cleanup() {
	e.m.Close()
}
func MGOSession() *mgo.Session {
	return e.m
}
func Port() string {
	return e.port
}
func MongoHost() string {
	return e.mongoHost
}
func MongoUser() string {
	return e.mongoUser
}
func MongoPwd() string {
	return e.mongoPwd
}
func MongoDb() string {
	return e.mongoDb
}
func SourcesCollection() string {
	return e.sourcesCollection
}
func LogPath() string {
	return e.logPath
}

func Init(configFile, configPath string) error {
	viper.SetConfigName(configFile)
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	mgoSession, err := mongo.NewClient(
		viper.GetString("mongo_host"),
		viper.GetString("mongo_user"),
		viper.GetString("mongo_pw"),
	)
	if err != nil {
		return fmt.Errorf("Error connecting to mongo: " + err.Error())
	}
	e.m = mgoSession

	e.port = viper.GetString("port")
	e.mongoHost = viper.GetString("mongo_host")
	e.mongoUser = viper.GetString("mongo_user")
	e.mongoPwd = viper.GetString("mongo_pwd")
	e.mongoDb = viper.GetString("mongo_db")
	e.sourcesCollection = viper.GetString("sources_collection")

	return nil
}
