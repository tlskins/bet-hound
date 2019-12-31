package main

import (
	"log"
	"net/http"
	"os"

	"bet-hound/cmd/env"
	"bet-hound/cmd/gql"
	m "bet-hound/pkg/mongo"

	// "github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/handler"
	// "github.com/go-chi/chi"
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

	mux := http.NewServeMux()

	corsOptions := cors.Options{
		AllowedHeaders: []string{
			"content-type",
			"authorization",
			"client-name",
			"client-version",
			"content-type",
		},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug:          false,
		AllowedOrigins: []string{"*"},
	}

	httpHandler := cors.New(corsOptions).Handler(mux)
	gqlConfig := gql.Config{Resolvers: &gql.Resolver{}}
	gqlOption := handler.WebsocketUpgrader(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	gqlHandler := handler.GraphQL(gql.NewExecutableSchema(gqlConfig), gqlOption)

	mux.Handle("/", handler.Playground("GraphQL playground", "/query"))
	mux.Handle("/query", gqlHandler)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	http.ListenAndServe(":"+port, httpHandler)
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
