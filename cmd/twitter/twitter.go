package main

import (
	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/nlp"
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

	// Initialize env
	err := env.Init(appConfigName, appConfigPath)
	if err != nil {
		logger.Fatalf("Error loading db config: %s \n", err)
	}
	defer env.Cleanup()
	m.Init(env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())
	fmt.Println("db config", env.MongoHost(), env.MongoUser(), env.MongoPwd(), env.MongoDb())

	//Register webhook
	if args := os.Args; len(args) > 1 && args[1] == "-register" {
		go registerWebhook(logger)
	}

	m := mux.NewRouter()
	m.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(200)
		fmt.Fprintf(writer, "Server is up and running")
	})
	m.HandleFunc("/webhook/twitter", CrcCheck).Methods("GET")
	m.HandleFunc("/webhook/twitter", WebhookHandler).Methods("POST")

	server := &http.Server{
		Handler: m,
	}
	server.Addr = ":9090"
	server.ListenAndServe()
}

func WebhookHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Handler called")
	logger.Println("Handler called")

	// Read and decode tweet
	body, _ := ioutil.ReadAll(request.Body)
	var load WebhookLoad
	err := json.Unmarshal(body, &load)
	if err != nil {
		fmt.Println("An error occured: " + err.Error())
		logger.Println("An error occured unmarshaling: " + err.Error())
	}

	//Check if it was a tweet_create_event and tweet was in the payload and it was not tweeted by the bot
	if len(load.TweetCreateEvent) < 1 || load.UserId == load.TweetCreateEvent[0].User.IdStr {
		return
	}

	fmt.Println("incoming created tweet", load.TweetCreateEvent[0])
	logger.Println("incoming created tweet", load.TweetCreateEvent[0])

	// Get tweet
	tweetId := load.TweetCreateEvent[0].IdStr
	url := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?tweet_mode=extended&id=%s", tweetId)
	client := CreateClient()
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("err", err)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	var tweet t.Tweet
	if err := json.Unmarshal([]byte(body), &tweet); err != nil {
		fmt.Println("err", err)
		panic(err)
	}
	fmt.Println("tweet data", tweet)
	logger.Println("tweet data", tweet)

	msg := tweet.FullText
	pHandle := "@" + tweet.User.ScreenName
	// inReplyToTweetId := tweet.InReplyToStatusIdStr
	// msg := data.(map[string]interface{})["full_text"]
	// msg := data["full_text"]
	// pHandle := data["user"]["screen_name"]
	// inReplyToTweetId := data["in_reply_to_status_id_str"]

	// msg := nlp.RemoveReservedTwitterWords(load.TweetCreateEvent[0].Text)
	// pHandle := "@" + load.TweetCreateEvent[0].User.Handle

	var response string
	bet, err := nlp.ParseNewText(msg, tweetId)
	if err != nil {
		fmt.Println("err parse new text", err)
		logger.Println("err parse new text", err)
		response = err.Error()
	} else {
		response = bet.Text()
		fmt.Println("created bet", response)
		db.UpsertBet(bet)
	}

	_, err = SendTweet(pHandle+" "+response, tweetId)
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
