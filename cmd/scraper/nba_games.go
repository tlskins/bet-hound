package scraper

import (
	"fmt"
	"time"

	gq "github.com/PuerkitoBio/goquery"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

var pbrSchedsRoot = "https://www.basketball-reference.com/leagues/NBA_2020_games-%s.htm"
var pbrGameRoot = "https://www.basketball-reference.com/boxscores/"
var pbrTz = "-0500 EST"
var pbrLoc = "America/New_York"

func ScrapeNbaGames() {
	fmt.Printf("%s: Scraping nba games\n", time.Now().String())
	months := []string{"february", "march", "april"}
	for _, month := range months {
		url := fmt.Sprintf(pbrSchedsRoot, month)
		doc, err := GetGqDocument(url)
		if err != nil {
			panic(err)
		}

		games := []*t.Game{}
		// Compile games
		doc.Find("#schedule > tbody > tr").Each(func(i int, s *gq.Selection) {
			boxScoreUri, _ := s.Find("[data-stat='box_score_text'] a").Attr("href")
			fk, _ := s.Find("[data-stat='date_game']").Attr("csk")
			// game has no box score so is not final
			if len(boxScoreUri) == 0 && len(fk) > 0 {
				dateStr := s.Find("[data-stat='date_game']").Text()
				timeStr := s.Find("[data-stat='game_start_time']").Text()
				gmTime, _ := time.Parse("Mon, Jan 2, 20063:04p-0700 MST", dateStr+timeStr+pbrTz)
				gameRes := GameResultTimeFor(&gmTime, pbrLoc)
				homeTmStr, _ := s.Find("[data-stat='home_team_name']").Attr("csk")
				homeTmFk := homeTmStr[0:3]
				homeTmNm := s.Find("[data-stat='home_team_name']").Text()
				awayTmStr, _ := s.Find("[data-stat='visitor_team_name']").Attr("csk")
				awayTmFk := awayTmStr[0:3]
				awayTmNm := s.Find("[data-stat='visitor_team_name']").Text()
				gm := t.Game{
					Id:            nbaLgId + fk,
					LeagueId:      nbaLgId,
					Fk:            fk,
					Url:           pbrGameRoot + fk,
					GameTime:      gmTime,
					GameResultsAt: gameRes,
					HomeTeamFk:    homeTmFk,
					HomeTeamName:  homeTmNm,
					AwayTeamFk:    awayTmFk,
					AwayTeamName:  awayTmNm,
				}
				fmt.Println(gm)
				games = append(games, &gm)
			}
		})

		// upsert
		if err := db.UpsertGames(&games); err != nil {
			panic(err)
		}
	}
}
