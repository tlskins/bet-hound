package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"bet-hound/cmd/cron"
	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/gql"
	"bet-hound/cmd/migration"
	"bet-hound/cmd/scraper"
	tw "bet-hound/cmd/twitter"
	m "bet-hound/pkg/mongo"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"

	// cron "github.com/robfig/cron/v3"
	"github.com/rs/cors"
)

const appConfigPath = "../../env"
const appConfigName = "config"

var logger *log.Logger

func main() {
	// Initialize env
	if err := env.Init(appConfigName, appConfigPath); err != nil {
		logger.Fatalf("Error loading db config: %s \n", err)
	}
	defer env.Cleanup()

	// init db
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())
	logger = SetUpLogger(env.LogPath(), env.LogName())
	mSess := env.MGOSession()
	db.EnsureIndexes(mSess.DB(env.MongoDb()))

	// cors & routers
	corsOptions := cors.Options{
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		// Debug:            true, // Enable Debugging for testing, consider disabling in production
		AllowedOrigins: strings.Split(env.AllowedOrigins(), ","),
	}
	corsHandler := cors.New(corsOptions).Handler
	router := chi.NewRouter()
	router.Use(corsHandler)

	// init graphql server
	gqlConfig := gql.New()
	gqlTimeout := handler.WebsocketKeepAliveDuration(10 * time.Second)
	gqlOption := handler.WebsocketUpgrader(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	gqlHandler := handler.GraphQL(gql.NewExecutableSchema(gqlConfig), gqlOption, gqlTimeout)
	gqlWithAuth := gql.AuthMiddleWare(gqlHandler, env.AppUrl())
	router.Handle("/query", gqlWithAuth)

	// init graphql playground
	plgWithAuth := gql.AuthMiddleWare(handler.Playground("GraphQL playground", "/query"), env.AppUrl())
	router.Handle("/playground", plgWithAuth)

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
	// cronSrv := cron.New(cron.WithLocation(env.TimeZone()))
	// if _, err := cronSrv.AddFunc("*/30 * * * *", ProcessRotoNfl(&gqlConfig)); err != nil {
	// 	fmt.Println(err)
	// }
	// cronSrv.Start()
	// defer cronSrv.Stop()

	// start graphql server
	go func() {
		log.Printf("connect to %s for GraphQL playground", env.GqlUrl())
		if err := http.ListenAndServe(":"+env.GqlPort(), router); err != nil {
			log.Fatal(err)
		}
	}()

	// options
	args := os.Args
	fmt.Println("args=", args)
	for _, arg := range args {
		if arg == "-scrape_nba_games" {
			fmt.Println("scraping nba games...")
			scraper.ScrapeNbaGames()
		} else if arg == "-seed_bet_maps" {
			fmt.Println("seeding bet maps...")
			migration.SeedBetMaps()
		} else if arg == "-disable_twitter" {
			fmt.Println("disabling twitter...")
			env.DisableTwitter()
		} else if arg == "-register" {
			fmt.Println("registering twitter webhook...")
			twtClient := env.TwitterClient()
			time.Sleep(5 * time.Second)
			twtClient.RegisterWebhook(env.WebhookEnv(), env.WebhookUrl())
		} else if arg == "-check_nba_games" {
			fmt.Println("checking nba games...")
			cron.CheckNbaGameResults(logger)
		}
	}

	select {} // block forever
}

func SetUpLogger(logPath, defaultPath string) *log.Logger {
	if logPath == "" {
		logPath = defaultPath
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(f, "", 0)
}

func ProcessRotoNfl(config *gql.Config) func() {
	return func() {
		r := config.Resolvers.Mutation()
		r.PostRotoArticle(context.Background())
	}
}
