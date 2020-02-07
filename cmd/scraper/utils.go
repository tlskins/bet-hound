package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	gq "github.com/PuerkitoBio/goquery"
)

var idRgx = regexp.MustCompile(`\/teams\/(.*)\/`)

func GetGqDocument(url string) (*gq.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	return gq.NewDocumentFromReader(res.Body)
}

func GameResultTimeFor(gameStart *time.Time, tzString string) time.Time {
	gmTimeTom := gameStart.AddDate(0, 0, 1)
	yrM, mthM, dayM := gmTimeTom.Date()
	loc, _ := time.LoadLocation(tzString)
	return time.Date(yrM, mthM, dayM, 9, 0, 0, 0, loc)
}

func TeamIdFor(teamUri string) (teamId string) {
	idMatch := idRgx.FindStringSubmatch(teamUri)
	if len(idMatch) > 1 {
		teamId = idMatch[1]
		teamId = strings.ToUpper(teamId)
	}
	return teamId
}

func EvaluateGameWinner(homeScore, awayScore float64) (homeWin, homeWinBy, homeLoseBy, awayWin, awayWinBy, awayLoseBy float64) {
	if homeScore == awayScore {
		homeWin = 0
		awayWin = 0
	} else if homeScore > awayScore {
		homeWin = 1
		awayWin = -1
	} else {
		homeWin = -1
		awayWin = 1
	}
	homeWinBy = homeScore - awayScore
	homeLoseBy = -1 * homeWinBy
	awayWinBy = awayScore - homeScore
	awayLoseBy = -1 * awayWinBy
	return
}
