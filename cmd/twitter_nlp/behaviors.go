package twitter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	b "bet-hound/cmd/betting"
	"bet-hound/cmd/db"
	"bet-hound/cmd/nlp"
	t "bet-hound/cmd/types"
)

func ProcessPendingFinalBets(twitterClient *http.Client) (err error) {
	bets := db.FindPendingFinalBets()
	for _, bet := range bets {
		fmt.Println("finalizing bet ", bet.String())
		err := b.CalcBetResult(bet)
		if err != nil {
			return err
		}

		respTweet, err := SendTweet(twitterClient, bet.Response(), bet.SourceFk)
		if err != nil {
			fmt.Println("err sending final bet tweet: ", err)
			return err
		}
		bet.BetResult.ResponseFk = respTweet.IdStr
		fmt.Println("Finalized bet: ", bet.Id, bet.Response())
		db.UpsertBet(bet)
	}
	return nil
}

func LoadTweet(twitterClient *http.Client, tweetId string) (tweet *t.Tweet, err error) {
	url := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?tweet_mode=extended&id=%s", tweetId)
	// client := CreateClient()
	resp, err := twitterClient.Get(url)
	if err != nil {
		return tweet, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	tweet = &t.Tweet{}
	err = json.Unmarshal([]byte(body), tweet)
	if err != nil {
		return tweet, err
	}
	err = db.UpsertTweet(tweet)
	return tweet, err
}

func SendTweet(twitterClient *http.Client, text string, replyId string) (responseTweet *t.Tweet, err error) {
	fmt.Println("Replying to tweet with", replyId, text)
	params := url.Values{}
	params.Set("status", text)
	params.Set("in_reply_to_status_id", replyId)

	resp, err := twitterClient.PostForm("https://api.twitter.com/1.1/statuses/update.json", params)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	responseTweet = &t.Tweet{}
	json.Unmarshal([]byte(body), responseTweet)
	if responseTweet == nil || responseTweet.IdStr == "" {
		var errResp *t.TwitterErrorResponse
		json.Unmarshal([]byte(body), errResp)
		if errResp != nil {
			for _, r := range errResp.Errors {
				fmt.Printf("Error sending tweet: %s %s\n", r.Code, r.Message)
			}
			return nil, fmt.Errorf(errResp.String())
		}
	}

	fmt.Println("Sent tweet to status: ", responseTweet.IdStr, responseTweet.InReplyToStatusIdStr, responseTweet.GetText())

	return responseTweet, nil
}

func ProcessNewTweet(twitterClient *http.Client, tweet *t.Tweet) (error, *t.Bet) {
	// Get full tweet
	tweetId := tweet.IdStr
	tweet, err := LoadTweet(twitterClient, tweetId)
	if err != nil {
		return err, nil
	}
	fmt.Println("tweet data", tweet)

	// Build Bet
	err, bet := b.BuildBetFromTweet(tweet)
	if err != nil {
		SendTweet(
			twitterClient,
			fmt.Sprintf("@%s %s", tweet.User.ScreenName, err.Error()),
			tweetId,
		)
	} else {
		responseTweet, err := SendTweet(twitterClient, bet.Response(), bet.SourceFk)
		if err != nil {
			return err, nil
		}
		bet.AcceptFk = responseTweet.IdStr
		err = db.UpsertBet(bet)
	}

	return err, bet
}

func ProcessReplyTweet(twitterClient *http.Client, tweet *t.Tweet, bet *t.Bet) (err error) {
	var yesRgx = regexp.MustCompile(`(?i)^(y(e|a)\S*|ok|sure|deal)`)
	var noRgx = regexp.MustCompile(`(?i)^(n(a|o)\S*|pass)`)
	text := strings.TrimSpace(nlp.RemoveReservedTwitterWords(tweet.GetText()))
	beforeStatus := bet.BetStatus.String()
	bet.PostProcess()

	// Process bet acceptance
	fmt.Printf("parse bet response text: %s %s\n", bet.Id, text)
	if yesRgx.Match([]byte(text)) {
		bet.AcceptBy(tweet.User.IdStr, tweet.IdStr)
	} else if noRgx.Match([]byte(text)) {
		bet.CancelBy(tweet.User.IdStr, tweet.IdStr)
	}

	// Respond if tweet changes bet status
	if beforeStatus != bet.BetStatus.String() {
		rTweet, err := SendTweet(twitterClient, bet.Response(), bet.SourceFk)
		if err != nil {
			return err
		}
		fmt.Println("Responded bet tweet id to id: ", rTweet.IdStr, bet.SourceFk, rTweet.GetText())
	}

	// Persist changes
	err = db.UpsertBet(bet)
	if err != nil {
		return err
	}

	return err
}
