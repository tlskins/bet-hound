package betting

import (
	// "bet-hound/cmd/db"
	"bet-hound/cmd/nlp"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
	"fmt"
	"github.com/satori/go.uuid"
	"math"
)

func BuildBetFromTweet(tweet *t.Tweet) (err error, bet *t.Bet) {
	err, eq := buildEquationFromText(*tweet.FullText)
	if err != nil {
		return err, nil
	}
	if len(tweet.Recipients()) == 0 {
		return fmt.Errorf("Not enough recipients!"), nil
	}
	recipient := tweet.Recipients()[0]
	bet = &t.Bet{
		Id:        uuid.NewV4().String(),
		SourceFk:  tweet.IdStr,
		Proposer:  tweet.User,
		Recipient: recipient,
		BetStatus: t.BetStatusFromString("Pending Proposer"),
		Equation:  *eq,
	}
	return nil, bet
}

func CalcBetResult(bet *t.Bet) (result string) {
	fmt.Println("calc bet result ", bet.Id, bet.Equation.Text())
	eq := bet.Equation
	games := scraper.ScrapeThisWeeksGames()
	lftMetric, rgtMetric := eq.Metrics()

	leftGame := t.FindGameByAwayFk(games, eq.LeftExpression.Player.TeamFk)
	if leftGame == nil {
		leftGame = t.FindGameByHomeFk(games, eq.LeftExpression.Player.TeamFk)
	}
	lftLog := scraper.ScrapeGameLog(leftGame)
	lftScore := calcPlayerGameScore(&lftLog, &eq.LeftExpression.Player, &lftMetric)

	rightGame := t.FindGameByAwayFk(games, eq.RightExpression.Player.TeamFk)
	if rightGame == nil {
		rightGame = t.FindGameByHomeFk(games, eq.RightExpression.Player.TeamFk)
	}
	rgtLog := scraper.ScrapeGameLog(rightGame)
	rgtScore := calcPlayerGameScore(&rgtLog, &eq.RightExpression.Player, &rgtMetric)

	var wId, lId, wPlayer, lPlayer string
	var wScore, lScore float64
	if lftScore > rgtScore {
		wId = bet.Proposer.ScreenName
		lId = bet.Recipient.ScreenName
		wPlayer = bet.Equation.LeftExpression.Player.Name
		lPlayer = bet.Equation.RightExpression.Player.Name
		wScore = lftScore
		lScore = rgtScore
	} else {
		wId = bet.Recipient.ScreenName
		lId = bet.Proposer.ScreenName
		wPlayer = bet.Equation.RightExpression.Player.Name
		lPlayer = bet.Equation.LeftExpression.Player.Name
		wScore = rgtScore
		lScore = lftScore
	}
	return fmt.Sprintf(
		"Congrats @%s you beat @%s! %s scored %.1f while %s only scored %.1f.",
		wId,
		lId,
		wPlayer,
		wScore,
		lPlayer,
		lScore,
	)
}

func calcPlayerGameScore(log *map[string]*t.GameStat, player *t.Player, metric *t.Metric) (score float64) {
	l := *log
	if l[player.Fk] == nil {
		fmt.Println(player.Fk, l)
	}
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
	return math.Ceil(score*10) / 10
}

func buildEquationFromText(text string) (err error, eq *t.Equation) {
	words := nlp.ParseText(text)
	opPhrase, leftMetric := nlp.FindOperatorPhrase(&words)
	leftPlayerExpr := nlp.FindLeftPlayerExpr(&words, opPhrase, leftMetric)
	fmt.Println("left player expr: ", leftPlayerExpr.Player.Name, leftPlayerExpr.Metric.Word.Text)
	rightPlayerExpr := nlp.FindRightPlayerExpr(&words, opPhrase, leftMetric)
	fmt.Println("right player expr: ", rightPlayerExpr.Player.Name)

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

	leftGame := t.FindGameByAwayFk(games, e.LeftExpression.Player.TeamFk)
	if leftGame == nil {
		leftGame = t.FindGameByHomeFk(games, e.LeftExpression.Player.TeamFk)
	}
	e.LeftExpression.Game = leftGame

	rightGame := t.FindGameByAwayFk(games, e.RightExpression.Player.TeamFk)
	if rightGame == nil {
		rightGame = t.FindGameByHomeFk(games, e.RightExpression.Player.TeamFk)
	}
	e.RightExpression.Game = rightGame
}
