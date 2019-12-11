package betting

import (
	// "bet-hound/cmd/db"
	"bet-hound/cmd/nlp"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
	"fmt"
	"github.com/satori/go.uuid"
	"math"
	"time"
)

func BuildBetFromTweet(tweet *t.Tweet) (err error, bet *t.Bet) {
	err, eq := BuildEquationFromText(*tweet.FullText)
	if err != nil {
		return err, nil
	}
	if len(tweet.Recipients()) == 0 {
		return fmt.Errorf("Not enough recipients!"), nil
	}
	recipient := tweet.Recipients()[0]
	var maxGmTime, minGmTime time.Time
	if eq.RightExpression.Game.GameTime.After(eq.LeftExpression.Game.GameTime) {
		maxGmTime = eq.RightExpression.Game.GameTime
		minGmTime = eq.LeftExpression.Game.GameTime
	} else {
		maxGmTime = eq.LeftExpression.Game.GameTime
		minGmTime = eq.RightExpression.Game.GameTime
	}

	loc, _ := time.LoadLocation("America/New_York")
	yrM, mthM, dayM := maxGmTime.Date()
	expiresAt := minGmTime.In(loc)
	// Toggle Expiration Here
	// if expiresAt.Before(time.Now()) {
	// 	return fmt.Errorf("Those games have already started."), nil
	// }
	// Toggle Expiration Here

	bet = &t.Bet{
		Id:          uuid.NewV4().String(),
		SourceFk:    tweet.IdStr,
		Proposer:    tweet.User,
		Recipient:   recipient,
		BetStatus:   t.BetStatusFromString("Pending Proposer"),
		Equation:    *eq,
		ExpiresAt:   expiresAt,
		FinalizedAt: time.Date(yrM, mthM, dayM, 9, 0, 0, 0, loc),
	}
	return nil, bet
}

func calcExpressionResult(expr *t.PlayerExpression, games *[]*t.Game, metric *t.Metric) (total float64, err error) {
	gm := t.FindGameByAwayFk(games, expr.Player.TeamFk)
	if gm == nil {
		gm = t.FindGameByHomeFk(games, expr.Player.TeamFk)
	}
	log := scraper.ScrapeGameLog(gm)
	score := calcPlayerGameScore(&log, &expr.Player, metric)

	if score == nil {
		return 0.0, fmt.Errorf("Unable to determine score for %s", expr.Description())
	}
	return *score, nil
}

func CalcBetResult(bet *t.Bet) (betRes *t.BetResult, err error) {
	fmt.Println("calc bet result ", bet.Id, bet.Text())
	eq := bet.Equation
	games := scraper.ScrapeThisWeeksGames()
	lftMetric, rgtMetric := eq.Metrics()

	// Calculate expression values
	leftRes, err := calcExpressionResult(&bet.Equation.LeftExpression, &games, &lftMetric)
	if err != nil {
		return nil, err
	}
	rightRes, err := calcExpressionResult(&bet.Equation.RightExpression, &games, &rgtMetric)
	if err != nil {
		return nil, err
	}

	var wSn, lSn string
	var wPlayer, lPlayer t.Player
	var wUsr, lUsr t.User
	var wScore, lScore float64
	if leftRes > rightRes {
		wUsr = bet.Proposer
		lUsr = bet.Recipient
		wSn = bet.Proposer.ScreenName
		lSn = bet.Recipient.ScreenName
		wPlayer = bet.Equation.LeftExpression.Player
		lPlayer = bet.Equation.RightExpression.Player
		wScore = leftRes
		lScore = rightRes
	} else {
		wUsr = bet.Recipient
		lUsr = bet.Proposer
		wSn = bet.Recipient.ScreenName
		lSn = bet.Proposer.ScreenName
		wPlayer = bet.Equation.RightExpression.Player
		lPlayer = bet.Equation.LeftExpression.Player
		wScore = rightRes
		lScore = leftRes
	}

	betRes = &t.BetResult{
		Winner:       wUsr,
		Loser:        lUsr,
		WinnerTotal:  wScore,
		LoserTotal:   lScore,
		Differential: wScore - lScore,
		Response: fmt.Sprintf("Congrats @%s you beat @%s! %s scored %.1f while %s only scored %.1f.",
			wSn,
			lSn,
			wPlayer.Name,
			wScore,
			lPlayer.Name,
			lScore,
		),
		DecidedAt: time.Now(),
	}
	return betRes, nil
}

func calcPlayerGameScore(log *map[string]*t.GameStat, player *t.Player, metric *t.Metric) *float64 {
	l := *log
	if l[player.Fk] == nil {
		fmt.Println("cant find game score", player.Fk, l)
		return nil
	}
	score := 0.0
	score += float64(l[player.Fk].PassYd) * 0.04
	score += float64(l[player.Fk].PassTd) * 4.0
	score -= float64(l[player.Fk].PassInt) * 2.0
	// score -= float64(l[player.Fk].PassSackedYd) / 10.0
	score += float64(l[player.Fk].RushYd) * 0.1
	score += float64(l[player.Fk].RushTd) * 6.0
	score += float64(l[player.Fk].Rec) * metric.PPR()
	score += float64(l[player.Fk].RecYd) * 0.1
	score += float64(l[player.Fk].RecTd) * 6.0
	score -= float64(l[player.Fk].FumbleLost) * 2.0
	score = math.Ceil(score*10) / 10
	return &score
}

func BuildEquationFromText(text string) (err error, eq *t.Equation) {
	words := nlp.ParseText(text)
	opPhrase, leftMetric := nlp.FindOperatorPhrase(&words)
	if opPhrase == nil {
		return fmt.Errorf("Sorry, couldn't find a betting operator (like 'score more than'!)"), nil
	}
	if leftMetric == nil {
		return fmt.Errorf("Sorry, couldn't find a betting metric (like 'ppr points')!"), nil
	}

	leftPlayerExpr := nlp.FindLeftPlayerExpr(&words, opPhrase, leftMetric)
	if leftPlayerExpr == nil {
		return fmt.Errorf("Sorry, couldn't a player for the proposer!"), nil
	}
	rightPlayerExpr := nlp.FindRightPlayerExpr(&words, opPhrase, leftMetric)
	if rightPlayerExpr == nil {
		return fmt.Errorf("Sorry, couldn't a player for the recipient!"), nil
	}

	eq = &t.Equation{
		LeftExpression:  *leftPlayerExpr,
		RightExpression: *rightPlayerExpr,
		Operator:        *opPhrase,
	}
	addGamesToEquation(eq)
	err = eq.Complete()

	return err, eq
}

func addGamesToEquation(e *t.Equation) {
	games := scraper.ScrapeThisWeeksGames()

	leftGame := t.FindGameByAwayFk(&games, e.LeftExpression.Player.TeamFk)
	if leftGame == nil {
		leftGame = t.FindGameByHomeFk(&games, e.LeftExpression.Player.TeamFk)
	}
	e.LeftExpression.Game = leftGame

	rightGame := t.FindGameByAwayFk(&games, e.RightExpression.Player.TeamFk)
	if rightGame == nil {
		rightGame = t.FindGameByHomeFk(&games, e.RightExpression.Player.TeamFk)
	}
	e.RightExpression.Game = rightGame
}
