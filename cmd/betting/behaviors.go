package betting

import (
	"bet-hound/cmd/db"
	"bet-hound/cmd/nlp"
	"bet-hound/cmd/scraper"
	t "bet-hound/cmd/types"
	"fmt"
	"github.com/satori/go.uuid"
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
		Id:               uuid.NewV4().String(),
		SourceFk:         tweet.IdStr,
		Proposer:         tweet.User,
		Recipient:        recipient,
		BetStatus:        t.BetStatusFromString("Pending Proposer"),
		ProposerCheckFk:  tweet.User.IdStr,
		RecipientCheckFk: recipient.IdStr,
		Equation:         *eq,
	}
	err = db.UpsertBet(bet)
	return err, bet
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
	if !eq.Complete() {
		err = fmt.Errorf("Incomplete equation!")
	}

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
