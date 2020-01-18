package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/gql"
	"bet-hound/cmd/gql/server/auth"
	"bet-hound/cmd/migration"
	tw "bet-hound/cmd/twitter"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	cron "github.com/robfig/cron/v3"
	"github.com/rs/cors"
)

const defaultPort = "8080"

const appConfigPath = "../../env"
const appConfigName = "config"

var logger *log.Logger

var lgSttgs *t.LeagueSettings

func main() {
	// Initialize
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	logger = SetUpLogger("", "logs.log")
	if err := env.Init(appConfigName, appConfigPath); err != nil {
		logger.Fatalf("Error loading db config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	// ensure indexes
	mSess := env.MGOSession()
	db.EnsureIndexes(mSess.DB(env.MongoDb()))

	corsOptions := cors.Options{
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		// Debug:            true, // Enable Debugging for testing, consider disabling in production
		AllowedOrigins: []string{"http://" + env.AppUrl() + ":3000", "http://" + env.AppUrl() + ":8080"},
	}
	corsHandler := cors.New(corsOptions).Handler
	router := chi.NewRouter()
	router.Use(corsHandler)

	// initialize graphql server
	gqlConfig := gql.New()
	gqlTimeout := handler.WebsocketKeepAliveDuration(10 * time.Second)
	gqlOption := handler.WebsocketUpgrader(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	gqlHandler := handler.GraphQL(gql.NewExecutableSchema(gqlConfig), gqlOption, gqlTimeout)
	router.Handle("/", auth.AuthMiddleWare(handler.Playground("GraphQL playground", "/query")))
	router.Handle("/query", auth.AuthMiddleWare(gqlHandler))

	// twitter server
	twt := env.TwitterClient()
	hookHandler := tw.WebhookHandlerWrapper(env.BotHandle())
	m := mux.NewRouter()
	m.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(200)
		fmt.Fprintf(writer, "Server is up and running")
	})
	m.HandleFunc("/webhook/twitter", tw.CrcCheck(env.ConsumerSecret())).Methods("GET")
	m.HandleFunc("/webhook/twitter", hookHandler(twt.Client)).Methods("POST")
	server := &http.Server{Handler: m, Addr: ":9090"}
	go server.ListenAndServe()
	fmt.Println("Twitter server running")

	// options
	if args := os.Args; len(args) > 1 {
		for _, arg := range args {
			if arg == "-register" {
				go twt.RegisterWebhook(env.WebhookEnv(), env.WebhookUrl())
			} else if arg == "-seed_users" {
				migration.SeedUsers()
			} else if arg == "-seed_nfl_players" {
				migration.SeedNflPlayers()
			} else if arg == "-seed_nfl_settings" {
				migration.SeedNflLeagueSettings()
			}
		}
	}

	// cron
	tz, err := time.LoadLocation(env.ServerTz())
	if err != nil {
		fmt.Println(err)
	}
	cronSrv := cron.New(cron.WithLocation(tz))
	lgSttgs = InitLeagueSettings(tz)
	lgSttgs.Print()
	if _, err := cronSrv.AddFunc("*/1 * * * *", ProcessEvents(lgSttgs, logger)); err != nil {
		fmt.Println(err)
	}
	if _, err := cronSrv.AddFunc("*/10 * * * *", ProcessRotoNfl(&gqlConfig)); err != nil {
		fmt.Println(err)
	}
	cronSrv.Start()
	defer cronSrv.Stop()

	// start graphql server
	log.Printf("connect to http://%s:%s/ for GraphQL playground", env.AppUrl(), port)
	http.ListenAndServe(":"+port, router)
	log.Fatal(http.ListenAndServe(":"+port, nil))
	fmt.Println("END")
}
