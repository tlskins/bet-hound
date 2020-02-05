package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
	utils "bet-hound/pkg/helpers"

	gq "github.com/PuerkitoBio/goquery"
)

func ScrapeNflTeams() error {
	fmt.Printf("%s: Scraping nfl teams...\n", time.Now().String())
	// Request the HTML page.
	res, err := http.Get("https://www.pro-football-reference.com/teams/")
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

	var teams []*t.Team
	nmsRgx := regexp.MustCompile(`(.+) (.+)$`)
	pfrUriRoot := "https://www.pro-football-reference.com"

	doc.Find("#teams_active > tbody > tr").Each(func(i int, s *gq.Selection) {
		tmA := s.Find("th[data-stat=team_name] a")
		if tmA != nil {
			var location, name string
			teamUri, _ := tmA.Attr("href")
			teamId := TeamIdFor(teamUri)
			fullName := tmA.Text()
			nmMatch := nmsRgx.FindStringSubmatch(fullName)
			if len(nmMatch) > 2 {
				location = nmMatch[1]
				name = nmMatch[2]
			}

			if len(teamId) > 0 {
				now := time.Now()
				team := &t.Team{
					Id:        teamId,
					LeagueId:  "nfl",
					Fk:        teamId,
					Url:       pfrUriRoot + teamUri,
					Name:      name,
					Location:  location,
					UpdatedAt: &now,
				}
				fmt.Println(utils.PrettyPrint(team))
				teams = append(teams, team)
			}
		}
	})

	if err = db.UpsertTeams(&teams); err != nil {
		fmt.Println(err)
	}

	return nil
}
