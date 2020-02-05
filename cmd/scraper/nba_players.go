package scraper

import (
	// t "bet-hound/cmd/types"
	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
	"fmt"
	"regexp"
	"strings"
	"time"

	gq "github.com/PuerkitoBio/goquery"
)

var pbrTeamRosterUrl = "https://www.basketball-reference.com/teams/%s/2020.html"
var pbrUri = "https://www.basketball-reference.com"
var pbrPlayerFkRgx *regexp.Regexp = regexp.MustCompile(".+\\/(.+).html")
var nbaLgId = "nba"

func ScrapeNbaPlayers() {
	fmt.Printf("%s: Scraping nba players...\n", time.Now().String())
	teamUrls := ScrapeNbaTeamUrls()
	players := []*t.Player{}
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
		doc, err = GetGqDocument(rosterUrl)
		if err != nil {
			panic(err)
		}
		teamName := doc.Find("#meta > div:nth-child(2) > h1 > span:nth-child(2)").Text()
		doc.Find("#roster > tbody > tr").Each(func(i int, s *gq.Selection) {
			fullName, _ := s.Find("[data-stat='player']").Attr("csk")
			names := strings.Split(fullName, ",")
			urlSuff, _ := s.Find("[data-stat='player'] a").Attr("href")
			fk := nbaPlayerFkFrom(urlSuff)
			pos := s.Find("[data-stat='pos']").Text()
			now := time.Now()
			player := t.Player{
				Id:        nbaLgId + fk,
				LeagueId:  nbaLgId,
				Fk:        fk,
				Name:      fmt.Sprintf("%s %s", names[1], names[0]),
				Url:       pbrUri + urlSuff,
				UpdatedAt: &now,
				FirstName: names[1],
				LastName:  names[0],
				TeamFk:    teamFk,
				TeamName:  teamName,
				Position:  pos,
			}
			fmt.Println(player)
			players = append(players, &player)
		})
	}

	if err := db.UpsertPlayers(&players); err != nil {
		panic(err)
	}
}

// helpers

func nbaPlayerFkFrom(text string) (fk string) {
	fkMatch := pbrPlayerFkRgx.FindStringSubmatch(text)
	if len(fkMatch) > 1 {
		fk = fkMatch[1]
		fk = strings.ToLower(fk)
	}
	return
}
