package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	gq "github.com/PuerkitoBio/goquery"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

var teamFkRgx *regexp.Regexp = regexp.MustCompile(`\/teams\/(.*)\/\d{4}\.htm`)

func ScrapeGameLog(url string) (gameLog *t.GameLog, err error) {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := gq.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	gameLog = &t.GameLog{}
	isFinal := false
	doc.Find("#content > div.linescore_wrap > table > tbody > tr").Each(func(i int, s *gq.Selection) {
		isFinal = true
		url, _ := s.Find("td:nth-child(2) a").Attr("href")
		fkMatch := teamFkRgx.FindStringSubmatch(url)
		fk := strings.ToUpper(fkMatch[1])
		name := s.Find("td:nth-child(2) a").Text()
		q1, _ := strconv.ParseFloat(s.Find("td:nth-child(3)").Text(), 64)
		q2, _ := strconv.ParseFloat(s.Find("td:nth-child(4)").Text(), 64)
		q3, _ := strconv.ParseFloat(s.Find("td:nth-child(5)").Text(), 64)
		q4, _ := strconv.ParseFloat(s.Find("td:nth-child(6)").Text(), 64)
		scrByQ := []float64{q1, q2, q3, q4}
		qf, _ := strconv.ParseFloat(s.Find("td:nth-child(7)").Text(), 64)

		tmLog := t.TeamLog{
			Fk:         fk,
			TeamName:   name,
			Score:      qf,
			ScoreByQtr: scrByQ,
		}

		if i == 0 {
			gameLog.AwayTeamLog = tmLog
		} else {
			gameLog.HomeTeamLog = tmLog
		}
	})

	if !isFinal {
		return nil, fmt.Errorf("Game not final")
	} else {
		if gameLog.HomeTeamLog.Score == gameLog.AwayTeamLog.Score {
			gameLog.HomeTeamLog.Win = 0
			gameLog.AwayTeamLog.Win = 0
		} else if gameLog.HomeTeamLog.Score > gameLog.AwayTeamLog.Score {
			gameLog.HomeTeamLog.Win = 1
			gameLog.AwayTeamLog.Win = -1
		} else {
			gameLog.HomeTeamLog.Win = -1
			gameLog.AwayTeamLog.Win = 1
		}
		gameLog.HomeTeamLog.WinBy = gameLog.HomeTeamLog.Score - gameLog.AwayTeamLog.Score
		gameLog.HomeTeamLog.LoseBy = -1 * gameLog.HomeTeamLog.WinBy
		gameLog.AwayTeamLog.WinBy = gameLog.AwayTeamLog.Score - gameLog.HomeTeamLog.Score
		gameLog.AwayTeamLog.LoseBy = -1 * gameLog.AwayTeamLog.WinBy
	}

	gameLog.PlayerLogs = scrapePlayerLogs(doc)

	return
}

func scrapePlayerLogs(doc *gq.Document) (playerLogs map[string]*t.PlayerLog) {
	playerLogs = make(map[string]*t.PlayerLog)

	doc.Find("#player_offense tbody").Each(func(i int, s *gq.Selection) {
		s.Find("tr").Each(func(i int, s *gq.Selection) {
			playerFk, _ := s.Find("th").Attr("data-append-csv")

			if len(playerFk) > 0 {
				log := t.PlayerLog{}
				s.Find("td").Each(func(i int, s *gq.Selection) {
					data, _ := s.Attr("data-stat")
					switch data {
					case "pass_cmp":
						log.PassCmp, _ = strconv.ParseFloat(s.Text(), 64)
					case "pass_att":
						log.PassAtt, _ = strconv.ParseFloat(s.Text(), 64)
					case "pass_yds":
						log.PassYd, _ = strconv.ParseFloat(s.Text(), 64)
					case "pass_td":
						log.PassTd, _ = strconv.ParseFloat(s.Text(), 64)
					case "pass_int":
						log.PassInt, _ = strconv.ParseFloat(s.Text(), 64)
					case "pass_sacked":
						log.PassSacked, _ = strconv.ParseFloat(s.Text(), 64)
					case "pass_sacked_yds":
						log.PassSackedYd, _ = strconv.ParseFloat(s.Text(), 64)
					case "pass_long":
						log.PassLong, _ = strconv.ParseFloat(s.Text(), 64)
					case "pass_rating":
						log.PassRating, _ = strconv.ParseFloat(s.Text(), 64)
					case "rush_att":
						log.RushAtt, _ = strconv.ParseFloat(s.Text(), 64)
					case "rush_yds":
						log.RushYd, _ = strconv.ParseFloat(s.Text(), 64)
					case "rush_td":
						log.RushTd, _ = strconv.ParseFloat(s.Text(), 64)
					case "rush_long":
						log.RushLong, _ = strconv.ParseFloat(s.Text(), 64)
					case "target":
						log.Target, _ = strconv.ParseFloat(s.Text(), 64)
					case "rec":
						log.Rec, _ = strconv.ParseFloat(s.Text(), 64)
					case "rec_yds":
						log.RecYd, _ = strconv.ParseFloat(s.Text(), 64)
					case "rec_td":
						log.RecTd, _ = strconv.ParseFloat(s.Text(), 64)
					case "rec_long":
						log.RecLong, _ = strconv.ParseFloat(s.Text(), 64)
					case "fumbles":
						log.Fumble, _ = strconv.ParseFloat(s.Text(), 64)
					case "fumbles_lost":
						log.FumbleLost, _ = strconv.ParseFloat(s.Text(), 64)
					}
				})
				log.CalcFantasyScores()
				playerLogs[playerFk] = &log
			}
		})
	})

	return
}

func ScrapeGames(gmYr, gmWk int) error {
	fmt.Printf("%s: Scraping games for year %d week %d\n", time.Now().String(), gmYr, gmWk)
	doc, err := GetWeeksGames(gmYr, gmWk)
	if err != nil {
		return err
	}

	games := []*t.Game{}
	pfrRoot := "https://www.pro-football-reference.com"
	gameUrls := make(map[string]string)

	// Compile games
	doc.Find(".game_summaries table.teams tbody").Each(func(i int, s *gq.Selection) {
		urlSuffix, _ := s.Find("tr td.gamelink a").Attr("href")
		url := strings.Join([]string{pfrRoot, urlSuffix}, "")
		var fkRgx = regexp.MustCompile(`\/boxscores\/(.*)\.htm`)
		fk := fkRgx.FindStringSubmatch(url)[1]

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

		now := time.Now()
		games = append(games, &t.Game{
			Id:           strconv.Itoa(gmYr) + strconv.Itoa(gmWk) + homeTeamFk + awayTeamFk,
			LeagueId:     "nfl",
			Name:         name,
			Fk:           fk,
			Url:          url,
			AwayTeamFk:   awayTeamFk,
			AwayTeamName: awayTeam,
			HomeTeamFk:   homeTeamFk,
			HomeTeamName: homeTeam,
			Final:        false,
			UpdatedAt:    &now,
		})
	})

	// Scrape game times
	gameTimes := make(map[string]*time.Time)
	for _, url := range gameUrls {
		if gameTimes[url] == nil {
			res, err := http.Get(url)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
			}
			doc, err := gq.NewDocumentFromReader(res.Body)
			if err != nil {
				return err
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
		gm.Week = gmWk
		gm.Year = gmYr
	}

	if err = db.UpsertGames(&games); err != nil {
		fmt.Println(err)
	}

	// Update league settings
	s := &t.LeagueSettings{
		Id:          "nfl",
		CurrentYear: gmYr,
		CurrentWeek: gmWk,
	}
	db.UpsertLeagueSettings(s)

	return nil
}

func GetWeeksGames(yr, wk int) (doc *gq.Document, err error) {
	res, err := http.Get(fmt.Sprintf("https://www.pro-football-reference.com/years/%d/week_%d.htm", yr, wk))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err = gq.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, err
}

func getThisWeeksGames() (doc *gq.Document, gmYr int, gmWk int, err error) {
	// Request this weeks games
	res, err := http.Get("https://www.pro-football-reference.com/boxscores/")
	if err != nil {
		return nil, 0, 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, 0, 0, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err = gq.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, 0, 0, err
	}

	// Toggle Expiration Here
	// return doc, err
	// Get week date
	gameDate := doc.Find("#content > div.section_heading > h2").Text()
	re := regexp.MustCompile(`^(?P<yr>\d+) Week (?P<wk>\d+)$`)
	match := re.FindStringSubmatch(gameDate)
	gmYrStr := ""
	if len(match) > 2 {

	}
	gmYrStr = match[1]
	gmYr, _ = strconv.Atoi(match[1])
	gmWk, _ = strconv.Atoi(match[2])
	nextGmWk := gmWk + 1
	gmWkStr := strconv.Itoa(nextGmWk)

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
		return doc, 0, 0, nil
	}

	// Pull next week if games final
	res, err = http.Get(fmt.Sprintf("https://www.pro-football-reference.com/years/%s/week_%s.htm", gmYrStr, gmWkStr))
	if err != nil {
		return nil, 0, 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, 0, 0, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err = gq.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, 0, 0, err
	}
	return doc, gmYr, nextGmWk, err
	// Toggle Expiration Here
}
