package betting

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"bet-hound/cmd/db"
	"bet-hound/cmd/env"
	"bet-hound/cmd/nlp"
	t "bet-hound/cmd/types"
)

func TweetBetProposal(bet *t.Bet) (*t.Tweet, error) {
	if bet.Recipient.TwitterUser == nil {
		return nil, fmt.Errorf("Bet recipient does not have a twitter account linked.")
	}
	// send tweet
	client := env.TwitterClient()
	betUrl := fmt.Sprintf("%s/bet/%s", env.AppUrl(), bet.Id)
	rcpt := bet.Recipient.TwitterUser.ScreenName
	prps := bet.Proposer.TwitterUser.ScreenName
	txt := fmt.Sprintf("@%s has proposed a bet to @%s %s. Do you accept? %s", prps, rcpt, bet.String(), betUrl)
	resp, err := client.SendTweet(txt, nil)
	if err != nil {
		return nil, err
	}
	// update bet
	bet.AcceptFk = "-1" // no resp or err means twitter
	if resp != nil {
		bet.AcceptFk = resp.IdStr
		if err = db.UpsertTweet(resp); err != nil {
			fmt.Println(err)
		}
	}
	if err = db.UpsertBet(bet); err != nil {
		return nil, err
	}
	return resp, nil
}

func ReplyToTweet(tweet *t.Tweet) error {
	// check if bet reply
	if tweet.InReplyToStatusIdStr != "" {
		if bet, err := db.FindBetByReply(tweet); err == nil && bet != nil {
			return replyToApproval(bet, tweet)
		}
	}
	text := strings.TrimSpace(nlp.RemoveReservedTwitterWords(tweet.GetText()))
	// check if user registration
	var registerRgx = regexp.MustCompile(`(?i)^register`)
	if registerRgx.Match([]byte(text)) {
		return replyToUserRegistration(tweet)
	}

	return nil
}

// private helpers

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func replyToUserRegistration(tweet *t.Tweet) error {
	userRgx := regexp.MustCompile(`(?i)^register[^ ]* +([^ ]+)`)
	text := strings.TrimSpace(nlp.RemoveReservedTwitterWords(tweet.GetText()))
	client := env.TwitterClient()

	userNameMatch := userRgx.FindStringSubmatch(text)
	if len(userNameMatch) < 2 {
		response := fmt.Sprintf("@%s Invalid user name", tweet.TwitterUser.ScreenName)
		if _, err := client.SendTweet(response, &tweet.IdStr); err != nil {
			return err
		}
		return fmt.Errorf(response)
	}

	userName := userNameMatch[1]
	if _, err := db.FindUserByUserName(userName); err == nil {
		response := fmt.Sprintf("@%s User name already exists", tweet.TwitterUser.ScreenName)
		if _, err := client.SendTweet(response, &tweet.IdStr); err != nil {
			return err
		}
		return fmt.Errorf(response)
	} else {
		usr, _ := db.FindUserByTwitterId(tweet.TwitterUser.IdStr)
		if len(usr.UserName) > 0 {
			response := fmt.Sprintf("@%s Already registered under username: %s", tweet.TwitterUser.ScreenName, usr.UserName)
			if _, err := client.SendTweet(response, &tweet.IdStr); err != nil {
				return err
			}
			return fmt.Errorf(response)
		} else {
			pwd := randString(8)
			if usr == nil {
				usr = &t.User{
					Name:        tweet.TwitterUser.Name,
					UserName:    userName,
					Password:    pwd,
					TwitterUser: &tweet.TwitterUser,
				}
			} else {
				// twitter user only exists
				usr.Password = pwd
				usr.Name = tweet.TwitterUser.Name
				usr.UserName = userName
				usr.TwitterUser = &tweet.TwitterUser
			}

			// register existing user
			if existing, _ := db.FindUserByTwitterScreenName(tweet.TwitterUser.ScreenName); existing != nil {
				usr.Id = existing.Id
			}

			response := fmt.Sprintf("You have been registered with username: %s. Your temporary password is: %s", userName, pwd)
			if _, err = client.SendDirectMessage(response, tweet.TwitterUser.IdStr); err != nil {
				return err
			}
			client.SendTweet(fmt.Sprintf("@%s Registration successful! Please check your DMs!", tweet.TwitterUser.ScreenName), &tweet.IdStr)
			_, err = db.UpsertUser(usr)
			return err
		}
	}
}

func replyToApproval(bet *t.Bet, tweet *t.Tweet) error {
	var yesRgx = regexp.MustCompile(`(?i)^(y(e|a)\S*|ok|sure|deal)`)
	var noRgx = regexp.MustCompile(`(?i)^(n(a|o)\S*|pass)`)
	text := strings.TrimSpace(nlp.RemoveReservedTwitterWords(tweet.GetText()))

	// process response
	if yesRgx.Match([]byte(text)) {
		twtUsr := tweet.TwitterUser
		rcpTwt := bet.Recipient.TwitterUser
		prpTwt := bet.Proposer.TwitterUser
		if (prpTwt.IdStr == twtUsr.IdStr) || (prpTwt.ScreenName == twtUsr.ScreenName) {
			bet.ProposerReplyFk = &tweet.IdStr
			if err := db.SyncTwitterUser(&twtUsr); err != nil {
				fmt.Println("error syncing: ", err)
			}
		} else if (rcpTwt.IdStr == twtUsr.IdStr) || (rcpTwt.ScreenName == twtUsr.ScreenName) {
			bet.RecipientReplyFk = &tweet.IdStr
			if err := db.SyncTwitterUser(&twtUsr); err != nil {
				fmt.Println("error syncing: ", err)
			}
		}
		if bet.ProposerReplyFk != nil && bet.RecipientReplyFk != nil {
			bet.BetStatus = t.BetStatusFromString("Accepted")
		}
	} else if noRgx.Match([]byte(text)) {
		bet.BetStatus = t.BetStatusFromString("Declined")
	}

	// reply to tweet
	txt := fmt.Sprintf("%s Bet status: %s", bet.TwitterHandles(), bet.BetStatus.String())
	if bet.BetStatus.String() == "Pending Approval" {
		pends := "recipient"
		if bet.ProposerReplyFk == nil {
			pends = "proposer and " + pends
		}
		txt = fmt.Sprintf("%s Pending approval from %s", bet.TwitterHandles(), pends)
	}
	client := env.TwitterClient()
	if resp, err := client.SendTweet(txt, &tweet.IdStr); err != nil {
		return nil
	} else if resp != nil {
		if err = db.UpsertTweet(resp); err != nil {
			fmt.Println(err)
		}
	}
	if err := db.UpsertBet(bet); err != nil {
		return err
	}
	return nil
}

// func BuildBetFromTweet(tweet *t.Tweet) (err error, bet *t.Bet) {
// 	text := strings.TrimSpace(nlp.RemoveReservedTwitterWords(tweet.GetText()))
// 	eqs, err := BuildEquationsFromText(text)
// 	if err != nil {
// 		return err, nil
// 	}
// 	if len(tweet.Recipients()) == 0 {
// 		return fmt.Errorf("Not enough recipients!"), nil
// 	}
// 	recipient := tweet.Recipients()[0]

// 	bet = &t.Bet{
// 		Id:        uuid.NewV4().String(),
// 		SourceFk:  tweet.IdStr,
// 		Proposer:  tweet.User,
// 		Recipient: recipient,
// 		BetStatus: t.BetStatusFromString("Pending Proposer"),
// 		Equations: eqs,
// 	}
// 	bet.PostProcess()
// 	valid := bet.Valid()
// 	return valid, bet
// }

// func BuildEquationsFromText(text string) (eqs []*t.Equation, err error) {
// 	allWords := nlp.ParseText(text)
// 	playerWords := nlp.FindPlayerWords(&allWords)
// 	currentGames := scraper.ScrapeThisWeeksGames()

// 	// Build Equations
// 	actionEqsMap := make(map[int]*t.Equation)
// 	for _, pw := range playerWords {
// 		// Find Player
// 		playerWord := pw[len(pw)-1]
// 		lemmas := nlp.WordsLemmas(&pw)
// 		player := db.SearchPlayerByName(strings.Join(lemmas, " "))
// 		if player == nil {
// 			fmt.Printf("Player not found.\n")
// 			continue
// 		}
// 		// Find game
// 		game := findGameByFk(&currentGames, player.TeamFk)
// 		if game == nil {
// 			fmt.Printf("Game not found for %s.\n", player.Name)
// 			continue
// 		}
// 		// Find action
// 		action := nlp.SearchLastParent(&allWords, playerWord.Index, -1, -1, []string{}, []string{"ACTION"})
// 		if action == nil {
// 			fmt.Printf("No action found for %s.\n", player.Name)
// 			continue
// 		}

// 		var eq *t.Equation
// 		var delimiter *t.Word
// 		// Get / Build equation
// 		if actionEqsMap[action.Index] != nil {
// 			eq = actionEqsMap[action.Index]
// 			delimiter = eq.Delimiter
// 		} else {
// 			delimiter = nlp.SearchFirstChild(&allWords, action.Index, -1, -1, []string{}, []string{"DELIMITER"})
// 			if delimiter == nil {
// 				fmt.Printf("No delimiter found for %s.\n", player.Name)
// 				continue
// 			}

// 			operator := nlp.SearchFirstChild(&allWords, action.Index, playerWord.Index, -1, []string{}, []string{"OPERATOR"})
// 			if operator == nil {
// 				fmt.Printf("No operator found for %s.\n", player.Name)
// 				continue
// 			}

// 			metricWord := nlp.SearchFirstChild(&allWords, action.Index, -1, -1, []string{}, []string{"METRIC"})
// 			var metric *t.Metric
// 			if metricWord == nil {
// 				fmt.Printf("No metric found for %s.\n", player.Name)
// 				continue
// 			} else {
// 				mods := nlp.SearchChildren(&allWords, metricWord.Index, -1, -1, []string{}, []string{"METRIC_MOD"})
// 				metric = &t.Metric{
// 					Word:      metricWord,
// 					Modifiers: mods,
// 				}
// 			}

// 			eq = &t.Equation{
// 				Action:    action,
// 				Metric:    metric,
// 				Delimiter: delimiter,
// 				Operator:  operator,
// 			}
// 			actionEqsMap[action.Index] = eq
// 		}

// 		// Build Expression
// 		expr := t.PlayerExpression{
// 			Player: player,
// 			Game:   game,
// 		}

// 		if playerWord.Index < delimiter.Index {
// 			eq.LeftExpressions = append(eq.LeftExpressions, &expr)
// 		} else {
// 			eq.RightExpressions = append(eq.RightExpressions, &expr)
// 		}
// 	}

// 	for _, eq := range actionEqsMap {
// 		eqs = append(eqs, eq)
// 	}
// 	if len(eqs) == 0 {
// 		return eqs, fmt.Errorf("No equations found!.")
// 	}
// 	return eqs, nil
// }
