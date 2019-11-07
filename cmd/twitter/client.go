package main

import (
	"bet-hound/cmd/env"
	"encoding/json"
	"fmt"
	// "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//Struct to parse webhook load
type WebhookLoad struct {
	UserId           string  `json:"for_user_id"`
	TweetCreateEvent []Tweet `json:"tweet_create_events"`
}

//Struct to parse tweet
type Tweet struct {
	Id    int64
	IdStr string `json:"id_str"`
	User  User
	Text  string
}

//Struct to parse user
type User struct {
	Id     int64
	IdStr  string `json:"id_str"`
	Name   string
	Handle string `json:"screen_name"`
}

func CreateClient() *http.Client {
	//Create oauth client with consumer keys and access token
	fmt.Println("create twitter client", env.ConsumerKey(), env.ConsumerSecret(), env.AccessTokenKey(), env.AccessTokenSecret())
	config := oauth1.NewConfig(env.ConsumerKey(), env.ConsumerSecret())
	token := oauth1.NewToken(env.AccessTokenKey(), env.AccessTokenSecret())
	httpClient := config.Client(oauth1.NoContext, token)

	return httpClient
}

func registerWebhook(logger *log.Logger) {
	logger.Println("Registering webhook...", env.WebhookEnv())
	fmt.Println("Registering webhook...")
	httpClient := CreateClient()

	//Set parameters
	path := "https://api.twitter.com/1.1/account_activity/all/" + env.WebhookEnv() + "/webhooks.json"
	hook_url := env.AppUrl() + "/webhook/twitter"
	logger.Println("path,hook_url", path, hook_url)
	values := url.Values{}
	values.Set("url", hook_url)

	//Make Oauth Post with parameters
	resp, err := httpClient.PostForm(path, values)
	if err != nil {
		logger.Println("httpClient.PostForm err", err)
	}
	fmt.Println("resp", resp)
	logger.Println("resp", resp)
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
	logger.Println("Webhook id of " + data["id"].(string) + " has been registered")
	subscribeWebhook()
}

func subscribeWebhook() {
	fmt.Println("Subscribing webapp...")
	client := CreateClient()
	path := "https://api.twitter.com/1.1/account_activity/all/" + env.WebhookEnv() + "/subscriptions.json"
	resp, _ := client.PostForm(path, nil)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	//If response code is 204 it was successful
	if resp.StatusCode == 204 {
		fmt.Println("Subscribed successfully")
	} else if resp.StatusCode != 204 {
		fmt.Println("Could not subscribe the webhook. Response below:")
		fmt.Println(string(body))
	}
}

func SendTweet(tweet string, reply_id string) (*Tweet, error) {
	fmt.Println("Sending tweet as reply to " + reply_id)
	//Initialize tweet object to store response in
	var responseTweet Tweet
	//Add params
	params := url.Values{}
	params.Set("status", tweet)
	params.Set("in_reply_to_status_id", reply_id)

	//Grab client and post
	client := CreateClient()
	resp, err := client.PostForm("https://api.twitter.com/1.1/statuses/update.json", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	//Decode response and send out
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	err = json.Unmarshal(body, &responseTweet)
	if err != nil {
		return nil, err
	}
	return &responseTweet, nil
}
