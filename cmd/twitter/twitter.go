package main

import (
	"bet-hound/cmd/twitter/env"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const appConfigPath = "./env"
const appConfigName = "config"

var logger *log.Logger

func main() {
	//Load env
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// 	fmt.Println("Error loading .env file")
	// }
	logger = setUpLogger("", "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading application config: %s \n", err)
	}

	fmt.Println("Starting Server")

	if args := os.Args; len(args) > 1 && args[1] == "-register" {
		go registerWebhook(logger)
	}
	//Create a new Mux Handler
	m := mux.NewRouter()
	//Listen to the base url and send a response
	m.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(200)
		fmt.Fprintf(writer, "Server is up and running")
	})
	//Listen to crc check and handle
	m.HandleFunc("/webhook/twitter", CrcCheck).Methods("GET")
	//Listen to webhook event and handle
	m.HandleFunc("/webhook/twitter", WebhookHandler).Methods("POST")

	//Start Server
	server := &http.Server{
		Handler: m,
	}
	server.Addr = ":9090"
	server.ListenAndServe()
}

func WebhookHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Handler called")
	logger.Println("Handler called")
	//Read the body of the tweet
	body, _ := ioutil.ReadAll(request.Body)
	//Initialize a webhok load obhject for json decoding
	var load WebhookLoad
	err := json.Unmarshal(body, &load)
	if err != nil {
		fmt.Println("An error occured: " + err.Error())
	}
	//Check if it was a tweet_create_event and tweet was in the payload and it was not tweeted by the bot
	if len(load.TweetCreateEvent) < 1 || load.UserId == load.TweetCreateEvent[0].User.IdStr {
		return
	}
	//Send Hello world as a reply to the tweet, replies need to begin with the handles
	//of accounts they are replying to
	_, err = SendTweet("@"+load.TweetCreateEvent[0].User.Handle+" Hello World", load.TweetCreateEvent[0].IdStr)
	if err != nil {
		fmt.Println("An error occured:")
		fmt.Println(err.Error())
	} else {
		fmt.Println("Tweet sent successfully")
	}
}

func CrcCheck(writer http.ResponseWriter, request *http.Request) {
	//Set response header to json type
	writer.Header().Set("Content-Type", "application/json")
	//Get crc token in parameter
	token := request.URL.Query()["crc_token"]
	if len(token) < 1 {
		fmt.Fprintf(writer, "No crc_token given")
		return
	}

	//Encrypt and encode in base 64 then return
	h := hmac.New(sha256.New, []byte(env.E.ConsumerSecret))
	h.Write([]byte(token[0]))
	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))
	//Generate response string map
	response := make(map[string]string)
	response["response_token"] = "sha256=" + encoded
	//Turn response map to json and send it to the writer
	responseJson, _ := json.Marshal(response)
	fmt.Fprintf(writer, string(responseJson))
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
