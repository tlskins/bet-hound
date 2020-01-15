package env

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/spf13/viper"

	tw "bet-hound/cmd/twitter/client"
	"bet-hound/pkg/mongo"
)

type environment struct {
	m  *mgo.Session
	tc *tw.TwitterClient

	consumerSecret           string
	webhookEnv               string
	appUrl                   string
	port                     string
	mongoHost                string
	mongoUser                string
	mongoPwd                 string
	mongoDb                  string
	betsCollection           string
	playersCollection        string
	tweetsCollection         string
	gamesCollection          string
	currentGamesCollection   string
	leagueSettingsCollection string
	usersCollection          string
	logPath                  string
	botHandle                string
}

var e = &environment{}

func ConsumerSecret() string {
	return e.consumerSecret
}
func WebhookEnv() string {
	return e.webhookEnv
}
func AppUrl() string {
	return e.appUrl
}
func Cleanup() {
	e.m.Close()
}
func TwitterClient() *tw.TwitterClient {
	return e.tc
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
func BetsCollection() string {
	return e.betsCollection
}
func PlayersCollection() string {
	return e.playersCollection
}
func TweetsCollection() string {
	return e.tweetsCollection
}
func GamesCollection() string {
	return e.gamesCollection
}
func CurrentGamesCollection() string {
	return e.currentGamesCollection
}
func LeagueSettingsCollection() string {
	return e.leagueSettingsCollection
}
func UsersCollection() string {
	return e.usersCollection
}
func LogPath() string {
	return e.logPath
}
func BotHandle() string {
	return e.botHandle
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

	e.tc = tw.CreateClient(
		viper.GetString("consumer_key"),
		viper.GetString("consumer_secret"),
		viper.GetString("access_token_key"),
		viper.GetString("access_token_secret"),
	)

	e.consumerSecret = viper.GetString("consumer_secret")
	e.webhookEnv = viper.GetString("webhook_env")
	e.appUrl = viper.GetString("app_url")
	e.port = viper.GetString("port")
	e.mongoHost = viper.GetString("mongo_host")
	e.mongoUser = viper.GetString("mongo_user")
	e.mongoPwd = viper.GetString("mongo_pwd")
	e.mongoDb = viper.GetString("mongo_db")
	e.betsCollection = viper.GetString("bets_collection")
	e.playersCollection = viper.GetString("players_collection")
	e.tweetsCollection = viper.GetString("tweets_collection")
	e.gamesCollection = viper.GetString("games_collection")
	e.currentGamesCollection = viper.GetString("current_games_collection")
	e.leagueSettingsCollection = viper.GetString("league_settings_collection")
	e.usersCollection = viper.GetString("users_collection")
	e.botHandle = viper.GetString("bot_handle")

	return nil
}
