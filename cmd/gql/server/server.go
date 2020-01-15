package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"bet-hound/cmd/env"
	"bet-hound/cmd/gql"
	"bet-hound/cmd/gql/server/auth"
	tw "bet-hound/cmd/twitter"
	m "bet-hound/pkg/mongo"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

const defaultPort = "8080"

const appConfigPath = "../../env"
const appConfigName = "config"

var logger *log.Logger

func main() {
	// Initialize
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	logger = setUpLogger("", "logs.log")
	if err := env.Init(appConfigName, appConfigPath); err != nil {
		logger.Fatalf("Error loading db config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	// graphql server
	corsOptions := cors.Options{
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true, // Enable Debugging for testing, consider disabling in production
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
	}
	corsHandler := cors.New(corsOptions).Handler
	router := chi.NewRouter()
	router.Use(corsHandler)

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
	if args := os.Args; len(args) > 1 && args[1] == "-register" {
		go twt.RegisterWebhook(env.WebhookEnv(), env.AppUrl())
	}
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

	// timed processes
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Timed processes at ", t)
				r := gqlConfig.Resolvers.Mutation()
				r.PostRotoArticle(context.Background())
			}
		}
	}()

	twt.SendTweet(fmt.Sprintf("@ckettstweets test %d", time.Now().Unix()), nil)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	http.ListenAndServe(":"+port, router)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func setUpLogger(logPath, defaultPath string) *log.Logger {
	if logPath == "" {
		logPath = defaultPath
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(f, "", 0)
}
