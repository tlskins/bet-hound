package twitter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	t "bet-hound/cmd/types"
)

func CrcCheck(consumerSecret string) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		token := request.URL.Query()["crc_token"]
		if len(token) < 1 {
			fmt.Fprintf(writer, "No crc_token given")
			return
		}

		h := hmac.New(sha256.New, []byte(consumerSecret))
		h.Write([]byte(token[0]))
		response := make(map[string]string)
		response["response_token"] = "sha256=" + base64.StdEncoding.EncodeToString(h.Sum(nil))
		responseJson, _ := json.Marshal(response)
		fmt.Fprintf(writer, string(responseJson))
	}
}

func WebhookHandlerWrapper(botHandle string) func(httpClient *http.Client) func(writer http.ResponseWriter, request *http.Request) {
	return func(httpClient *http.Client) func(writer http.ResponseWriter, request *http.Request) {
		return func(writer http.ResponseWriter, request *http.Request) {
			fmt.Println("Handler called")
			body, _ := ioutil.ReadAll(request.Body)
			var load t.WebhookLoad
			if err := json.Unmarshal(body, &load); err != nil {
				fmt.Println(err)
			}

			if len(load.TweetCreateEvent) < 1 || len(load.TweetCreateEvent[0].IdStr) == 0 || load.TweetCreateEvent[0].TwitterUser.ScreenName == botHandle {
				fmt.Println("filtered out tweet")
				return
			}

			newTweet := load.TweetCreateEvent[0]
			fmt.Println("incoming tweet text, id, replyTo: ", newTweet.GetText(), newTweet.IdStr, newTweet.InReplyToStatusIdStr)
			// var bet *t.Bet
			// // Check if response to a check tweet
			// if len(newTweet.InReplyToStatusIdStr) > 0 {
			// 	bet = db.FindBetByReply(&newTweet)
			// }

			// Reply to proposer check
			// if bet != nil {
			// 	logger.Println("processing reply to bet", bet.Id, bet.Description())
			// 	err = twitter.ProcessReplyTweet(httpClient, &newTweet, bet)
			// 	if err != nil {
			// 		logger.Println("err processing reply tweet", err)
			// 	}
			// } else {
			// 	// Process a new bet
			// 	logger.Println("processing new tweet...")
			// 	err, bet := twitter.ProcessNewTweet(httpClient, &newTweet)
			// 	if err != nil {
			// 		logger.Println("err processing new tweet", err)
			// 	} else {
			// 		logger.Println("created bet: ", bet.Id)
			// 	}
			// }

			// if err != nil {
			// 	fmt.Println("An error occured:")
			// 	fmt.Println(err.Error())
			// } else {
			// 	fmt.Println("Tweet handled successfully")
			// }
		}
	}
}
