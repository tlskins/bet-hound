package main

import (
	"bet-hound/cmd/scraper"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/gql"
	mw "bet-hound/cmd/gql/server/middleware"
	"bet-hound/cmd/migration"
	tw "bet-hound/cmd/twitter"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	cron "github.com/robfig/cron/v3"
	"github.com/rs/cors"
)

const appConfigPath = "../../env"
const appConfigName = "config"

var logger *log.Logger
var lgSttgs *t.LeagueSettings

func main() {
	// Initialize
	if err := env.Init(appConfigName, appConfigPath); err != nil {
		logger.Fatalf("Error loading db config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())
	logger = SetUpLogger(env.LogPath(), env.LogName())
	mSess := env.MGOSession()
	db.EnsureIndexes(mSess.DB(env.MongoDb()))

	corsOptions := cors.Options{
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		// Debug:            true, // Enable Debugging for testing, consider disabling in production
		AllowedOrigins: strings.Split(env.AllowedOrigins(), ","),
	}
	corsHandler := cors.New(corsOptions).Handler
	router := chi.NewRouter()
	router.Use(corsHandler)

	// seed and settings options
	args := os.Args
	fmt.Println("args=", args)
	for _, arg := range args {
		if arg == "-seed_users" {
			fmt.Println("seeding users...")
			migration.SeedUsers()
		} else if arg == "-seed_nfl_players" {
			fmt.Println("seeding nfl players...")
			migration.SeedNflPlayers()
		} else if arg == "-seed_nfl_teams" {
			fmt.Println("seeding nfl teams...")
			migration.SeedNflTeams()
		} else if arg == "-seed_nfl_settings" {
			fmt.Println("seeding nfl settings...")
			migration.SeedNflLeagueSettings()
		} else if arg == "-seed_nfl_curr_gms" {
			fmt.Println("seeding current games...")
			scraper.ScrapeGames(2019, 20)
		} else if arg == "-disable_twitter" {
			fmt.Println("disabling twitter...")
			env.DisableTwitter()
		}
	}

	// initialize league settings
	tz, err := time.LoadLocation(env.ServerTz())
	if err != nil {
		panic(err)
	}
	lgSttgs = InitLeagueSettings("nfl")
	// lgSttgs.Print()

	// init graphql server
	gqlConfig := gql.New()
	gqlTimeout := handler.WebsocketKeepAliveDuration(10 * time.Second)
	gqlOption := handler.WebsocketUpgrader(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	gqlHandler := handler.GraphQL(gql.NewExecutableSchema(gqlConfig), gqlOption, gqlTimeout)
	gqlWithAuth := mw.AuthMiddleWare(gqlHandler, env.AppUrl())
	gqlWithLg := mw.LeagueMiddleWare(gqlWithAuth, lgSttgs)
	router.Handle("/query", gqlWithLg)

	// init graphql playground
	plgWithAuth := mw.AuthMiddleWare(handler.Playground("GraphQL playground", "/query"), env.AppUrl())
	plgWithLg := mw.LeagueMiddleWare(plgWithAuth, lgSttgs)
	router.Handle("/playground", plgWithLg)

	// init twitter server
	router.Get("/webhook/twitter", tw.CrcCheck(env.ConsumerSecret()))
	router.Post("/webhook/twitter", tw.WebhookHandlerWrapper(env.BotHandle()))
	router.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(200)
		fmt.Fprintf(writer, "Server is up and running")
	})

	// health check
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// cron
	cronSrv := cron.New(cron.WithLocation(tz))
	if _, err := cronSrv.AddFunc("*/30 * * * *", ProcessEvents(lgSttgs, logger)); err != nil {
		fmt.Println(err)
	}
	if _, err := cronSrv.AddFunc("*/30 * * * *", ProcessRotoNfl(&gqlConfig)); err != nil {
		fmt.Println(err)
	}
	cronSrv.Start()
	defer cronSrv.Stop()

	// start graphql server
	go func() {
		log.Printf("connect to %s for GraphQL playground", env.GqlUrl())
		if err := http.ListenAndServe(":"+env.GqlPort(), router); err != nil {
			log.Fatal(err)
		}
	}()

	// server options
	for _, arg := range args {
		if arg == "-register" {
			twtClient := env.TwitterClient()
			time.Sleep(5 * time.Second)
			twtClient.RegisterWebhook(env.WebhookEnv(), env.WebhookUrl())
		} else if arg == "-process_events" {
			ProcessEvents(lgSttgs, logger)()
		}
	}

	select {} // block forever
}
