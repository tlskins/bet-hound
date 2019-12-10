package main

import (
	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/twitter"
	t "bet-hound/cmd/types"
	m "bet-hound/pkg/mongo"
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

const appConfigPath = "../env"
const appConfigName = "config"

var logger *log.Logger

func main() {
	// Initialize
	fmt.Println("Starting Server")
	logger = setUpLogger("", "logs.log")
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading db config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())
	fmt.Println("db config", env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	// Create client
	client := CreateClient()

	// Register webhook
	if args := os.Args; len(args) > 1 && args[1] == "-register" {
		go registerWebhook(client, logger)
	}

	// Process pending final bets
	twitter.ProcessPendingFinalBets(client)

	// Setup handler
	m := mux.NewRouter()
	m.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(200)
		fmt.Fprintf(writer, "Server is up and running")
	})
	m.HandleFunc("/webhook/twitter", CrcCheck).Methods("GET")
	m.HandleFunc("/webhook/twitter", WebhookHandlerWrapper(client)).Methods("POST")

	server := &http.Server{
		Handler: m,
	}
	server.Addr = ":9090"
	server.ListenAndServe()
}

func WebhookHandlerWrapper(httpClient *http.Client) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Println("Handler called")

		// Read and decode tweet
		body, _ := ioutil.ReadAll(request.Body)
		var load WebhookLoad
		err := json.Unmarshal(body, &load)
		if err != nil {
			logger.Println("An error occured unmarshaling: " + err.Error())
		}

		//Check if it was a tweet_create_event and tweet was in the payload and it was not tweeted by the bot
		if len(load.TweetCreateEvent) < 1 || len(load.TweetCreateEvent[0].IdStr) == 0 || load.UserId == load.TweetCreateEvent[0].User.IdStr {
			logger.Println("filtered out tweet")
			return
		}

		newTweet := load.TweetCreateEvent[0]
		logger.Println("incoming created tweet", newTweet.GetText(), newTweet.IdStr)

		// Check if response to a check tweet
		replyTweetId := newTweet.InReplyToStatusIdStr
		logger.Println("replyTweetId", replyTweetId)
		var bet *t.Bet
		if len(replyTweetId) > 0 {
			bet, err = db.FindBetByReply(&newTweet)
			// TODO : Send reply tweet that its invalid bet if start time < now and expire bet
			if err != nil {
				fmt.Println("FindBetByReply err ", err.Error())
			}
		}

		// Reply to proposer check
		if bet != nil && err == nil {
			logger.Println("reply to bet", bet.Id, bet.Equation.Text())
			err = twitter.ProcessReplyTweet(httpClient, &newTweet, bet)
			if err != nil {
				logger.Println("err processing reply tweet", err)
			}
		} else {
			// Process a new bet
			logger.Println("processing new tweet...")
			err, bet := twitter.ProcessNewTweet(httpClient, &newTweet)
			if err != nil {
				logger.Println("err processing new tweet", err)
			} else {
				logger.Println("created bet: ", bet.Id)
			}
		}

		if err != nil {
			fmt.Println("An error occured:")
			fmt.Println(err.Error())
		} else {
			fmt.Println("Tweet handled successfully")
		}
	}
}

// func WebhookHandler(writer http.ResponseWriter, request *http.Request) {
// 	logger.Println("Handler called")

// 	// Read and decode tweet
// 	body, _ := ioutil.ReadAll(request.Body)
// 	var load WebhookLoad
// 	err := json.Unmarshal(body, &load)
// 	if err != nil {
// 		logger.Println("An error occured unmarshaling: " + err.Error())
// 	}

// 	//Check if it was a tweet_create_event and tweet was in the payload and it was not tweeted by the bot
// 	if len(load.TweetCreateEvent) < 1 || load.UserId == load.TweetCreateEvent[0].User.IdStr {
// 		logger.Println("filtered out tweet: ", len(load.TweetCreateEvent), load.UserId, load.TweetCreateEvent[0].User.IdStr, load.TweetCreateEvent[0])
// 		return
// 	}

// 	newTweet := load.TweetCreateEvent[0]
// 	logger.Println("incoming created tweet", newTweet.GetText(), newTweet.User.IdStr)

// 	// Check if response to a check tweet
// 	replyTweetId := newTweet.InReplyToStatusIdStr
// 	logger.Println("replyTweetId", replyTweetId)
// 	var bet *t.Bet
// 	if len(replyTweetId) > 0 {
// 		bet, err = db.FindBetByReply(&newTweet)
// 		if err != nil {
// 			fmt.Println("FindBetByReply err ", err.Error())
// 		}
// 	}

// 	// Reply to proposer check
// 	if bet != nil && err == nil {
// 		logger.Println("reply to bet", bet.Id, bet.Equation.Text())
// 		err = twitter.ProcessReplyTweet(client, &newTweet, bet)
// 		if err != nil {
// 			logger.Println("err processing reply tweet", err)
// 		}
// 	} else {
// 		// Process a new bet
// 		logger.Println("processing new tweet...")
// 		err, bet := twitter.ProcessNewTweet(client, &newTweet)
// 		if err != nil {
// 			logger.Println("err processing new tweet", err)
// 		} else {
// 			logger.Println("created bet: ", bet.Id)
// 		}
// 	}

// 	if err != nil {
// 		fmt.Println("An error occured:")
// 		fmt.Println(err.Error())
// 	} else {
// 		fmt.Println("Tweet handled successfully")
// 	}
// }

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
	h := hmac.New(sha256.New, []byte(env.ConsumerSecret()))
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
