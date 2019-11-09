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
	"net/url"
	"os"
	"regexp"
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

func LoadTweet(tweetId string) (tweet *t.Tweet, err error) {
	url := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?tweet_mode=extended&id=%s", tweetId)
	client := CreateClient()
	resp, err := client.Get(url)
	if err != nil {
		return tweet, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	tweet = &t.Tweet{}
	err = json.Unmarshal([]byte(body), tweet)
	return tweet, err
}

func WebhookHandler(writer http.ResponseWriter, request *http.Request) {
	logger.Println("Handler called")

	// Read and decode tweet
	body, _ := ioutil.ReadAll(request.Body)
	var load WebhookLoad
	err := json.Unmarshal(body, &load)
	if err != nil {
		logger.Println("An error occured unmarshaling: " + err.Error())
	}

	//Check if it was a tweet_create_event and tweet was in the payload and it was not tweeted by the bot
	if len(load.TweetCreateEvent) < 1 || load.UserId == load.TweetCreateEvent[0].User.IdStr {
		logger.Println("filtered out tweet: ", len(load.TweetCreateEvent), load.UserId, load.TweetCreateEvent[0].User.IdStr, load.TweetCreateEvent[0])
		return
	}

	newTweet := load.TweetCreateEvent[0]
	logger.Println("incoming created tweet", newTweet.GetText(), newTweet.User.IdStr)

	// Check if response to a check tweet
	replyTweetId := newTweet.InReplyToStatusIdStr
	logger.Println("replyTweetId", replyTweetId)
	var bet *t.Bet
	if len(replyTweetId) > 0 {
		bet, err = db.FindBetByReply(&newTweet)
		if err != nil {
			fmt.Println("FindBetByReply err ", err.Error())
		}
	}

	// Reply to proposer check
	if bet != nil && err == nil {
		logger.Println("reply to bet", bet.Id, bet.Text())
		err = ProcessReplyTweet(&newTweet, bet)
		if err != nil {
			logger.Println("err processing reply tweet", err)
			panic(err)
		}
	} else {
		// Process a new bet
		logger.Println("processing new tweet...")
		err = ProcessNewTweet(&newTweet)
		if err != nil {
			logger.Println("err processing new tweet", err)
			panic(err)
		}
	}

	if err != nil {
		fmt.Println("An error occured:")
		fmt.Println(err.Error())
	} else {
		fmt.Println("Tweet handled successfully")
	}
}

func ProcessReplyTweet(tweet *t.Tweet, bet *t.Bet) (err error) {
	var yesRgx = regexp.MustCompile(`(?i)yes`)
	text := tweet.GetText()
	logger.Println("process reply text: ", text)
	if yesRgx.Match([]byte(text)) {
		if bet.BetStatus.String() == "Pending Proposer" {
			bet.BetStatus = t.BetStatusFromString("Pending Recipient")
			responseTweet, err := SendTweet(bet.Response(), *bet.Fk)
			if err != nil {
				return err
			}
			bet.RecipientCheckTweetId = &responseTweet.IdStr
			_, err = db.UpsertBet(bet)
			logger.Println("Sent check to recipient")
		} else if bet.BetStatus.String() == "Pending Recipient" {
			bet.BetStatus = t.BetStatusFromString("Accepted")
			responseTweet, err := SendTweet(bet.Response(), *bet.Fk)
			if err != nil {
				return err
			}
			bet.RecipientCheckTweetId = &responseTweet.IdStr
			_, err = db.UpsertBet(bet)
			logger.Println("Bet accepted")
		}
	} else {
		logger.Println("Did not reply yes")
	}
	return err
}

func ProcessNewTweet(tweet *t.Tweet) error {
	// Get full tweet
	tweetId := tweet.IdStr
	tweet, err := LoadTweet(tweetId)
	if err != nil {
		logger.Println("err loading tweet", err)
		panic(err)
	}
	fmt.Println("tweet data", tweet)
	logger.Println("tweet data", tweet)

	// Build Bet
	bet, err := nlp.ParseTweet(tweet)
	if err != nil {
		logger.Println("err parsing tweet", err)
		panic(err)
	}
	logger.Println("process new tweet created bet id", bet.Id)

	responseTweet, err := SendTweet(bet.Response(), tweetId)
	bet.ProposerCheckTweetId = &responseTweet.IdStr
	_, err = db.UpsertBet(bet)
	return err
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

func SendTweet(text string, replyId string) (responseTweet *t.Tweet, err error) {
	fmt.Println("Sending tweet as reply to " + replyId)
	logger.Println("Sending tweet as reply to " + replyId)
	params := url.Values{}
	params.Set("status", text)
	params.Set("in_reply_to_status_id", replyId)

	//Grab client and post
	client := CreateClient()
	resp, err := client.PostForm("https://api.twitter.com/1.1/statuses/update.json", params)
	if err != nil {
		logger.Println("err sending tweet", err)
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	responseTweet = &t.Tweet{}
	if err = json.Unmarshal([]byte(body), responseTweet); err != nil {
		logger.Println("err unmarshalling responseTweet", err)
		return nil, err
	}
	logger.Println("Sent tweet "+responseTweet.IdStr, responseTweet.GetText())
	return responseTweet, nil
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
