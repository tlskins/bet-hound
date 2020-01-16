package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"

	t "bet-hound/cmd/types"
)

type TwitterClient struct {
	Client *http.Client
}

func CreateClient(consumerKey, consumerSecret, accessKey, accessSecret string) *TwitterClient {
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessKey, accessSecret)
	return &TwitterClient{Client: config.Client(oauth1.NoContext, token)}
}

func (c *TwitterClient) SendTweet(text string, replyId *string) (responseTweet *t.Tweet, err error) {
	fmt.Println("SendTweet: ", text)
	params := url.Values{}
	params.Set("status", text)
	if replyId != nil {
		params.Set("in_reply_to_status_id", *replyId)
	}
	resp, err := c.Client.PostForm("https://api.twitter.com/1.1/statuses/update.json", params)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	responseTweet = &t.Tweet{}
	json.Unmarshal([]byte(body), responseTweet)
	if responseTweet == nil || responseTweet.IdStr == "" {
		fmt.Println("twitter err response: ", string(body))
	}
	fmt.Println("Sent tweet: ", responseTweet.IdStr, responseTweet.InReplyToStatusIdStr, responseTweet.GetText())

	return responseTweet, nil
}

func (c *TwitterClient) RegisterWebhook(webhookEnv, webhookUrl string) {
	fmt.Println("Registering webhook...")

	//Set parameters
	path := fmt.Sprintf("https://api.twitter.com/1.1/account_activity/all/%s/webhooks.json", webhookEnv)
	hook_url := fmt.Sprintf("%s/webhook/twitter", webhookUrl)
	values := url.Values{}
	values.Set("url", hook_url)

	//Make Oauth Post with parameters
	resp, err := c.Client.PostForm(path, values)
	if err != nil {
		fmt.Println("httpClient.PostForm err", err)
	}
	fmt.Println("resp", resp)
	defer resp.Body.Close()

	//Parse response and check response
	body, _ := ioutil.ReadAll(resp.Body)
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		fmt.Println("err", err)
		panic(err)
	}
	fmt.Println("data", data)
	fmt.Println("Webhook id of " + data["id"].(string) + " has been registered")
	subscribeWebhook(webhookEnv, c.Client)
}

func subscribeWebhook(webhookEnv string, client *http.Client) {
	fmt.Println("Subscribing webapp...")
	path := fmt.Sprintf("https://api.twitter.com/1.1/account_activity/all/%s/subscriptions.json", webhookEnv)
	resp, _ := client.PostForm(path, nil)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		fmt.Println("Subscribed successfully")
	} else if resp.StatusCode != 204 {
		fmt.Printf("Could not subscribe the webhook: %s", string(body))
	}
}
