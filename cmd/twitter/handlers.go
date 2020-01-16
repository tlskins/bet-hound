package twitter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	b "bet-hound/cmd/betting"
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
			fmt.Println("WebhookHandlerWrapper called")
			body, _ := ioutil.ReadAll(request.Body)
			var load t.WebhookLoad
			if err := json.Unmarshal(body, &load); err != nil {
				fmt.Println(err)
			}

			if len(load.TweetCreateEvent) < 1 {
				return
			}
			newTweet := load.TweetCreateEvent[0]
			if len(newTweet.IdStr) == 0 || newTweet.TwitterUser.ScreenName == botHandle {
				return
			}

			fmt.Println("incoming tweet text, id, replyTo: ", newTweet.GetText(), newTweet.IdStr, newTweet.InReplyToStatusIdStr)
			if err := b.ReplyToTweet(&newTweet); err != nil {
				fmt.Println(err)
			}
		}
	}
}
