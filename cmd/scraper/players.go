package scraper

import (
	"fmt"
	gq "github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"strings"
	"time"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

func ScrapePlayers() error {
	fmt.Printf("%s: Scraping players...\n", time.Now().String())
	// Request the HTML page.
	res, err := http.Get("https://www.pro-football-reference.com/years/2019/fantasy.htm")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := gq.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	var players []*t.Player

	doc.Find("#fantasy tr").Each(func(i int, s *gq.Selection) {
		headTd := s.Find("td[data-stat=player]")
		name := headTd.Text()
		names := strings.SplitN(name, " ", 2)
		var lastName string
		firstName := names[0]
		if len(names) > 1 {
			lastName = names[1]
		}
		id, _ := headTd.Attr("data-append-csv")
		url, _ := headTd.Find("a").Attr("href")

		teamA := s.Find("td[data-stat=team] a")
		teamName, _ := teamA.Attr("title")
		teamShort := teamA.Text()

		position := s.Find("td[data-stat=fantasy_pos]").Text()

		if len(id) > 0 {
			idRgx := regexp.MustCompile(`\/teams\/(.*)\/\d{4}\.htm`)
			teamUri, _ := teamA.Attr("href")
			var teamId string
			if len(idRgx.FindStringSubmatch(teamUri)) > 1 {
				teamId = idRgx.FindStringSubmatch(teamUri)[1]
				teamId = strings.ToUpper(teamId)
			}

			fmt.Printf("Player %d: %s %s %s %s %s %s\n", i, name, id, teamId, teamName, position, url)
			players = append(players, &t.Player{
				Id:        id,
				LeagueId:  "nfl",
				Name:      name,
				FirstName: firstName,
				LastName:  lastName,
				Fk:        id,
				TeamFk:    teamId,
				TeamName:  teamName,
				TeamShort: teamShort,
				Position:  position,
				Url:       url,
			})
		}
	})
	if err = db.UpsertPlayers(&players); err != nil {
		return err
	}
	return nil
}
