package twitter

import (
	b "bet-hound/cmd/betting"
	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

func ProcessPendingFinalBets(twitterClient *http.Client) (err error) {
	bets := db.FindPendingFinal()
	for _, bet := range bets {
		result := b.CalcBetResult(bet)
		bet.Result = result
		bet.BetStatus = t.BetStatusFinal
		_, err := SendTweet(twitterClient, result, bet.SourceFk)
		if err != nil {
			fmt.Println("err sending final bet tweet: ", err)
			return err
		}
		fmt.Println("Final bet id: ", bet.Id, b.CalcBetResult(bet))
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
	fmt.Println("Sending tweet ", replyId, text)
	params := url.Values{}
	params.Set("status", text)
	params.Set("in_reply_to_status_id", replyId)

	resp, err := twitterClient.PostForm("https://api.twitter.com/1.1/statuses/update.json", params)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	responseTweet = &t.Tweet{}
	if err = json.Unmarshal([]byte(body), responseTweet); err != nil {
		return nil, err
	}
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
		SendTweet(twitterClient, err.Error(), tweetId)
	} else {
		responseTweet, err := SendTweet(twitterClient, bet.Response(), tweetId)
		if err != nil {
			return err, nil
		}
		bet.ProposerCheckFk = responseTweet.IdStr
		err = db.UpsertBet(bet)
	}

	return err, bet
}

func ProcessReplyTweet(twitterClient *http.Client, tweet *t.Tweet, bet *t.Bet) (err error) {
	var yesRgx = regexp.MustCompile(`(?i)yes`)
	text := tweet.GetText()
	if yesRgx.Match([]byte(text)) {
		if bet.BetStatus.String() == "Pending Proposer" {
			bet.BetStatus = t.BetStatusPendingRecipient
			fmt.Println("Sending response tweet to recipient: ", bet.Response(), tweet.IdStr)
			responseTweet, err := SendTweet(twitterClient, bet.Response(), tweet.IdStr)
			if err != nil {
				return err
			}
			bet.RecipientCheckFk = responseTweet.IdStr
			err = db.UpsertBet(bet)
			if err != nil {
				fmt.Println("err upserting bet after proposer reply:", bet)
			}
		} else if bet.BetStatus.String() == "Pending Recipient" {
			bet.BetStatus = t.BetStatusAccepted
			fmt.Println("Sending response bet accepted: ", bet.Response(), tweet.IdStr)
			_, err := SendTweet(twitterClient, bet.Response(), tweet.IdStr)
			if err != nil {
				return err
			}
			err = db.UpsertBet(bet)
			if err != nil {
				fmt.Println("err upserting bet after proposer reply:", bet)
			}
		}
	} else {
		// logger.Println("Did not reply yes")
	}
	return err
}
