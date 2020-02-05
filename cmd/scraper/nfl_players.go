package scraper

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	gq "github.com/PuerkitoBio/goquery"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

var pfrPlayersUrl = "https://www.pro-football-reference.com/years/2019/fantasy.htm"

func ScrapeNflPlayers() error {
	fmt.Printf("%s: Scraping nfl players...\n", time.Now().String())
	doc, err := GetGqDocument(pfrPlayersUrl)
	if err != nil {
		panic(err)
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
			now := time.Now()
			players = append(players, &t.Player{
				Id:        id,
				LeagueId:  "nfl",
				Name:      name,
				FirstName: firstName,
				LastName:  lastName,
				Fk:        id,
				TeamFk:    teamId,
				TeamName:  teamName,
				Position:  position,
				Url:       url,
				UpdatedAt: &now,
			})
		}
	})
	if err = db.UpsertPlayers(&players); err != nil {
		return err
	}
	return nil
}
