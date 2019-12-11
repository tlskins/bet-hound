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
	"time"
)

func ProcessPendingFinalBets(twitterClient *http.Client) (err error) {
	bets := db.FindPendingFinal()
	for _, bet := range bets {
		fmt.Println("finalizing bet ", bet.Text())
		result := b.CalcBetResult(bet)
		if result == nil {
			continue
		}
		bet.Result = *result
		bet.BetStatus = t.BetStatusFinal
		_, err := SendTweet(twitterClient, bet.Result, bet.SourceFk)
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
	if err = json.Unmarshal([]byte(body), responseTweet); err != nil {
		fmt.Println("err sending tweet: ", err)
		return nil, err
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
	var yesRgx = regexp.MustCompile(`(?i)yes`)
	text := tweet.GetText()

	if bet.BetStatus.String() != "Pending Approval" {
		SendTweet(twitterClient, bet.Response(), tweet.IdStr)
		return fmt.Errorf("Bet status no longer pending: %s", bet.BetStatus.String())
	} else if bet.ExpiresAt.Before(time.Now()) {
		bet.BetStatus = t.BetStatusExpired
		SendTweet(twitterClient, bet.Response(), tweet.IdStr)
		err = db.UpsertBet(bet)
		if err != nil {
			return err
		}
		return fmt.Errorf("Bet expired")
	}

	if yesRgx.Match([]byte(text)) {
		if bet.ProposerReplyFk == nil && bet.Proposer.IdStr == tweet.User.IdStr {
			bet.ProposerReplyFk = &tweet.InReplyToStatusIdStr
		} else if bet.RecipientReplyFk == nil && bet.Recipient.IdStr == tweet.User.IdStr {
			bet.RecipientReplyFk = &tweet.InReplyToStatusIdStr
		}

		if bet.RecipientReplyFk != nil && bet.ProposerReplyFk != nil {
			bet.BetStatus = t.BetStatusAccepted
			rTweet, err := SendTweet(twitterClient, bet.Response(), bet.SourceFk)
			fmt.Println("Sent response tweet id to id: ", rTweet.IdStr, bet.SourceFk, rTweet.GetText())
			if err != nil {
				return err
			}
		}

		err = db.UpsertBet(bet)
		if err != nil {
			return err
		}
	}

	return err
}
