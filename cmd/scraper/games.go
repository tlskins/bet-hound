package scraper

import (
	gq "github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	// "bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

func ScrapeThisWeeksGames() (games []*t.Game) {
	// Request the HTML page.
	res, err := http.Get("https://www.pro-football-reference.com/boxscores/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := gq.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	pfrRoot := "https://www.pro-football-reference.com"
	gameDayTxt := doc.Find(".game_summaries div:nth-child(1) table.teams tbody tr:nth-child(1) td").Text()

	doc.Find(".game_summaries table.teams tbody").Each(func(i int, s *gq.Selection) {
		urlSuffix, _ := s.Find("tr td.gamelink a").Attr("href")
		url := strings.Join([]string{pfrRoot, urlSuffix}, "")
		var fkRgx = regexp.MustCompile(`\/boxscores\/(.*)\.htm`)
		fk := fkRgx.FindStringSubmatch(url)[1]

		gameTimeTxt := s.Find("tr:nth-child(3) td:nth-child(3)").Text()
		gameTimeTxt = strings.TrimSpace(gameTimeTxt)
		gameDateTxt := strings.Join([]string{gameDayTxt, gameTimeTxt}, " ")
		gameTime, _ := time.Parse("Jan 2, 2006 3:04pm", gameDateTxt)

		var teamFkRgx = regexp.MustCompile(`\/teams\/(.*)\/\d{4}\.htm`)
		// Away team fields
		awayTeam := s.Find("tr:nth-child(2) td:nth-child(1) a").Text()
		awayTeamUrl, _ := s.Find("tr:nth-child(2) td:nth-child(1) a").Attr("href")
		awayTeamMatch := teamFkRgx.FindStringSubmatch(awayTeamUrl)
		awayTeamFk := strings.ToUpper(awayTeamMatch[1])

		// Home team fields
		homeTeam := s.Find("tr:nth-child(3) td:nth-child(1) a").Text()
		homeTeamUrl, _ := s.Find("tr:nth-child(3) td:nth-child(1) a").Attr("href")
		homeTeamMatch := teamFkRgx.FindStringSubmatch(homeTeamUrl)
		homeTeamFk := strings.ToUpper(homeTeamMatch[1])

		name := strings.Join([]string{awayTeam, homeTeam}, " at ")
		name = strings.Join([]string{name, gameTime.Format("Jan 2, 2006 3:04pm")}, " ")

		// game already happened
		// if len(gameTimeTxt) > 0 {
		games = append(games, &t.Game{
			Name:         &name,
			Fk:           &fk,
			Url:          &url,
			AwayTeamFk:   &awayTeamFk,
			AwayTeamName: &awayTeam,
			HomeTeamFk:   &homeTeamFk,
			HomeTeamName: &homeTeam,
			GameTime:     &gameTime,
		})
		// }
	})

	return games
}
