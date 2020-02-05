package scraper

import (
	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
	"fmt"
	"regexp"
	"strings"
	"time"

	gq "github.com/PuerkitoBio/goquery"
)

var pbrTeamsUrl = "https://www.basketball-reference.com/teams/"
var pbrTeamUrl = "https://www.basketball-reference.com/teams/%s/"
var pbrLocRgx *regexp.Regexp = regexp.MustCompile("Location:[\\s]+([a-zA-Z ]+),")
var pbrTeamFkRgx *regexp.Regexp = regexp.MustCompile("\\/teams\\/(.+)\\/")

func ScrapeNbaTeams() {
	fmt.Printf("%s: Scraping nba teams...\n", time.Now().String())
	teamUrls := ScrapeNbaTeamUrls()
	teams := []*t.Team{}
	for _, teamUrl := range teamUrls {
		fmt.Println("scraping team url ", teamUrl)
		doc, err := GetGqDocument(teamUrl)
		if err != nil {
			panic(err)
		}

		teamUrlFk := nbaTeamFkFrom(teamUrl)
		teamUrlSuff, _ := doc.Find(fmt.Sprintf("#%s > tbody > tr:nth-child(1) > td[data-stat='team_name'] > a", teamUrlFk)).Attr("href")
		teamFk := nbaTeamFkFrom(teamUrlSuff)
		rosterUrl := pbrUri + teamUrlSuff
		loc := nbaLocationFrom(doc.Find("#meta > div:nth-child(2) > p:nth-child(2)").Text())
		doc.Find("#meta > div:nth-child(2)").Each(func(i int, s *gq.Selection) {
			fullName := s.Find("h1 span").Text()
			// loc := nbaLocationFrom(s.Find("p:nth-child(2)").Text())
			name := strings.Replace(fullName, loc, "", 1)
			name = strings.TrimSpace(name)
			now := time.Now()
			team := t.Team{
				Id:        nbaLgId + teamFk,
				LeagueId:  nbaLgId,
				Fk:        teamFk,
				Name:      name,
				Url:       rosterUrl,
				UpdatedAt: &now,
				Location:  loc,
			}
			fmt.Println(team)
			teams = append(teams, &team)
		})
	}

	if err := db.UpsertTeams(&teams); err != nil {
		panic(err)
	}
}

// helpers

func ScrapeNbaTeamUrls() (teamUrls []string) {
	fmt.Printf("%s: Scraping nba team FKs...\n", time.Now().String())
	doc, err := GetGqDocument(pbrTeamsUrl)
	if err != nil {
		panic(err)
	}

	// get team ids from teams url
	doc.Find("#teams_active > tbody > tr.full_table").Each(func(i int, s *gq.Selection) {
		teamUrl, _ := s.Find("[data-stat='franch_name'] a").Attr("href")
		teamUrls = append(teamUrls, pbrUri+teamUrl)
	})

	return
}

func nbaLocationFrom(text string) (loc string) {
	locMatch := pbrLocRgx.FindStringSubmatch(text)
	if len(locMatch) > 1 {
		loc = locMatch[1]
	}
	return
}

func nbaTeamFkFrom(text string) (fk string) {
	teamFkMatch := pbrTeamFkRgx.FindStringSubmatch(text)
	if len(teamFkMatch) > 1 {
		fk = teamFkMatch[1]
	}
	return
}
