package scraper

import (
	"fmt"
	gq "github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	// "bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

func ScrapeGameLog(game *t.Game) (gameLog map[string]*t.GameStat) {
	// Request the HTML page.
	res, err := http.Get(game.Url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := gq.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	gameLog = make(map[string]*t.GameStat)

	doc.Find("#player_offense tbody").Each(func(i int, s *gq.Selection) {
		s.Find("tr").Each(func(i int, s *gq.Selection) {
			playerFk, _ := s.Find("th").Attr("data-append-csv")

			if len(playerFk) > 0 {
				stat := t.GameStat{PlayerFk: playerFk}
				s.Find("td").Each(func(i int, s *gq.Selection) {
					data, _ := s.Attr("data-stat")
					switch data {
					case "pass_cmp":
						stat.PassCmp, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "pass_att":
						stat.PassAtt, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "pass_yds":
						stat.PassYd, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "pass_td":
						stat.PassTd, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "pass_int":
						stat.PassInt, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "pass_sacked":
						stat.PassSacked, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "pass_sacked_yds":
						stat.PassSackedYd, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "pass_long":
						stat.PassLong, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "pass_rating":
						stat.PassRating, _ = strconv.ParseFloat(s.Text(), 64)
					case "rush_att":
						stat.RushAtt, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "rush_yds":
						stat.RushYd, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "rush_td":
						stat.RushTd, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "rush_long":
						stat.RushLong, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "target":
						stat.Target, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "rec":
						stat.Rec, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "rec_yds":
						stat.RecYd, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "rec_td":
						stat.RecTd, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "rec_long":
						stat.RecLong, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "fumbles":
						stat.Fumble, _ = strconv.ParseInt(s.Text(), 0, 64)
					case "fumbles_lost":
						stat.FumbleLost, _ = strconv.ParseInt(s.Text(), 0, 64)
					}
				})
				gameLog[stat.PlayerFk] = &stat
			}
		})
	})
	return gameLog
}

func ScrapeThisWeeksGames() (games []*t.Game) {
	doc, err := getThisWeeksGames()
	if err != nil {
		log.Fatal(err)
	}

	pfrRoot := "https://www.pro-football-reference.com"
	gameUrls := make(map[string]string)

	// Compile games
	doc.Find(".game_summaries table.teams tbody").Each(func(i int, s *gq.Selection) {
		urlSuffix, _ := s.Find("tr td.gamelink a").Attr("href")
		url := strings.Join([]string{pfrRoot, urlSuffix}, "")
		var fkRgx = regexp.MustCompile(`\/boxscores\/(.*)\.htm`)
		fk := fkRgx.FindStringSubmatch(url)[1]

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

		// Update map for game urls to pull dates from
		gameUrls[homeTeamFk] = url
		gameUrls[awayTeamFk] = url
		name := strings.Join([]string{awayTeam, homeTeam}, " at ")

		games = append(games, &t.Game{
			Name:         name,
			Fk:           fk,
			Url:          url,
			AwayTeamFk:   awayTeamFk,
			AwayTeamName: awayTeam,
			HomeTeamFk:   homeTeamFk,
			HomeTeamName: homeTeam,
		})
	})

	// Scrape game times
	gameTimes := make(map[string]*time.Time)
	for _, url := range gameUrls {
		if gameTimes[url] == nil {
			// Request the HTML page.
			res, err := http.Get(url)
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

			doc.Find(".scorebox_meta").Each(func(i int, s *gq.Selection) {
				gmTime := fmt.Sprintf(
					"%s%s %s",
					s.Find("div:nth-child(1)").Text(),
					strings.Replace(s.Find("div:nth-child(2)").Text(), "Start Time:", "", -1),
					" -0500 EST", // This will depend on the server time zone i guess which is what the browser will render
				)
				date, err := time.Parse("Monday Jan 2, 2006 3:04pm -0700 MST", gmTime)
				if err != nil {
					fmt.Println(err)
				}
				gameTimes[url] = &date
			})
		}
	}

	// Add game times
	for _, gm := range games {
		gm.GameTime = *gameTimes[gm.Url]
		date := gm.GameTime.AddDate(0, 0, 1)
		yrM, mthM, dayM := date.Date()
		loc, _ := time.LoadLocation("America/New_York")
		// Results at game date + 1 day @ 9AM EST
		gm.GameResultsAt = time.Date(yrM, mthM, dayM, 9, 0, 0, 0, loc)
		if gm.GameTime.Before(time.Now()) {
			gm.Final = true
		} else {
			gm.Final = false
		}
	}

	return games
}

func getThisWeeksGames() (*gq.Document, error) {
	// Request this weeks games
	res, err := http.Get("https://www.pro-football-reference.com/boxscores/")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return nil, err
	}
	doc, err := gq.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Toggle Expiration Here
	// return doc, err
	// Get week date
	var gmYr, gmWk string
	gameDate := doc.Find("#content > div.section_heading > h2").Text()
	re := regexp.MustCompile(`^(?P<yr>\d+) Week (?P<wk>\d+)$`)
	match := re.FindStringSubmatch(gameDate)
	gmYr = match[1]
	gmWkInt, _ := strconv.ParseInt(match[2], 0, 64)
	gmWkInt += 1
	gmWk = strconv.FormatInt(gmWkInt, 10)

	// Check if all games finalized
	notFinal := false
	doc.Find(".game_summaries table.teams tbody").Each(func(i int, s *gq.Selection) {
		txt := s.Find("tr td.gamelink a").Text()
		if txt != "Final" {
			notFinal = true
			return
		}
	})
	if notFinal {
		return doc, nil
	}

	// Pull next week if games final
	res, err = http.Get(fmt.Sprintf("https://www.pro-football-reference.com/years/%s/week_%s.htm", gmYr, gmWk))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return nil, err
	}
	doc, err = gq.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return doc, err
	// Toggle Expiration Here
}
