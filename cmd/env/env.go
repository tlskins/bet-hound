package env

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/spf13/viper"

	tw "bet-hound/cmd/twitter/client"
	"bet-hound/pkg/mongo"
)

type environment struct {
	m        *mgo.Session
	tc       *tw.TwitterClient
	timeZone *time.Location

	consumerSecret         string
	webhookEnv             string
	webhookUrl             string
	appUrl                 string
	appHost                string
	appPort                string
	gqlUrl                 string
	gqlPort                string
	twitterPort            string
	mongoHost              string
	mongoUser              string
	mongoPwd               string
	mongoDb                string
	betsCollection         string
	playersCollection      string
	tweetsCollection       string
	gamesCollection        string
	teamsCollection        string
	usersCollection        string
	betMapsCollection      string
	leaderBoardsCollection string
	leagueStart            string
	leagueStart2           string
	leagueEnd              string
	leagueLastWeek         int
	logName                string
	logPath                string
	botHandle              string
	allowedOrigins         string
	serverTz               string
	awsSesAccessKeyId      string
	awsSesSecretAccessKey  string
}

var e = &environment{}

func ConsumerSecret() string {
	return e.consumerSecret
}
func WebhookEnv() string {
	return e.webhookEnv
}
func WebhookUrl() string {
	return e.webhookUrl
}
func AppUrl() string {
	return e.appUrl
}
func AppHost() string {
	return e.appHost
}
func AppPort() string {
	return e.appPort
}
func GqlUrl() string {
	return e.gqlUrl
}
func GqlPort() string {
	return e.gqlPort
}
func TwitterPort() string {
	return e.twitterPort
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
func UsersCollection() string {
	return e.usersCollection
}
func BetMapsCollection() string {
	return e.betMapsCollection
}
func TeamsCollection() string {
	return e.teamsCollection
}
func LeaderBoardsCollection() string {
	return e.leaderBoardsCollection
}
func LeagueStart() string {
	return e.leagueStart
}
func LeagueStart2() string {
	return e.leagueStart2
}
func LeagueEnd() string {
	return e.leagueEnd
}
func LeagueLastWeek() int {
	return e.leagueLastWeek
}
func LogName() string {
	return e.logName
}
func LogPath() string {
	return e.logPath
}
func BotHandle() string {
	return e.botHandle
}
func TimeZone() *time.Location {
	return e.timeZone
}
func AllowedOrigins() string {
	return e.allowedOrigins
}
func DisableTwitter() {
	e.tc.Disabled = true
}
func ServerTz() string {
	return e.serverTz
}
func AwsSesAccessKeyId() string {
	return e.awsSesAccessKeyId
}
func AwsSesSecretAccessKey() string {
	return e.awsSesSecretAccessKey
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
	e.webhookUrl = viper.GetString("webhook_url")
	e.appUrl = viper.GetString("app_url")
	e.appHost = viper.GetString("app_host")
	e.appPort = viper.GetString("app_port")
	e.gqlUrl = viper.GetString("gql_url")
	e.gqlPort = viper.GetString("gql_port")
	e.twitterPort = viper.GetString("twitter_port")
	e.mongoHost = viper.GetString("mongo_host")
	e.mongoUser = viper.GetString("mongo_user")
	e.mongoPwd = viper.GetString("mongo_pwd")
	e.mongoDb = viper.GetString("mongo_db")
	e.betsCollection = viper.GetString("bets_collection")
	e.playersCollection = viper.GetString("players_collection")
	e.tweetsCollection = viper.GetString("tweets_collection")
	e.gamesCollection = viper.GetString("games_collection")
	e.usersCollection = viper.GetString("users_collection")
	e.betMapsCollection = viper.GetString("bet_maps_collection")
	e.teamsCollection = viper.GetString("teams_collection")
	e.leaderBoardsCollection = viper.GetString("leader_boards_collection")
	e.botHandle = viper.GetString("bot_handle")
	e.leagueStart = viper.GetString("league_start")
	e.leagueStart2 = viper.GetString("league_start2")
	e.leagueEnd = viper.GetString("league_end")
	e.leagueLastWeek = viper.GetInt("league_last_week")
	e.logName = viper.GetString("log_name")
	e.logPath = viper.GetString("log_path")
	e.allowedOrigins = viper.GetString("allowed_origins")
	e.serverTz = viper.GetString("server_tz")
	if e.timeZone, err = time.LoadLocation(e.serverTz); err != nil {
		return err
	}
	e.awsSesAccessKeyId = viper.GetString("aws_ses_access_key_id")
	e.awsSesSecretAccessKey = viper.GetString("aws_ses_secret_access_key")

	return nil
}
