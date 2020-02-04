package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	gq "github.com/PuerkitoBio/goquery"

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

// needs to be refactored for schedules
// func ScrapeNflGames(year int) error {
// 	fmt.Printf("%s: Scraping nfl games\n", time.Now().String())
// 	scheduleUrl := fmt.Sprintf("https://www.pro-football-reference.com/years/%d/games.htm", year)
// 	res, err := http.Get(scheduleUrl)
// 	if err != nil {
// 		return err
// 	}
// 	defer res.Body.Close()
// 	if res.StatusCode != 200 {
// 		return fmt.Errorf("scraping status code error: %d %s", res.StatusCode, res.Status)
// 	}
// 	doc, err := gq.NewDocumentFromReader(res.Body)
// 	if err != nil {
// 		return err
// 	}

// 	// Compile games
// 	doc.Find("#games > tbody > tr").Each(func(i int, s *gq.Selection) {
// 		boxScoreUri, _ := s.Find("[data-stat='boxscore_word']").Attr("href")
// 		fk, _ := s.Find("[data-stat='boxscore_word'] a").Attr("href")

// 		// game has no box score so is not final
// 		if len(boxScoreUri) != 0 && len(fk) > 0 {
// 			dateStr := s.Find("[data-stat='date_game']").Text()
// 			timeStr := s.Find("[data-stat='game_start_time']").Text()
// 			gmTime, _ := time.Parse("Mon, Jan 2, 20063:04p-0700 MST", dateStr+timeStr+pbrTz)
// 			gameRes := GameResultTimeFor(&gmTime)
// 			homeTmStr, _ := s.Find("[data-stat='home_team_name']").Attr("csk")
// 			homeTmFk := homeTmStr[0:3]
// 			homeTmNm := s.Find("[data-stat='home_team_name']").Text()
// 			awayTmStr, _ := s.Find("[data-stat='away_team_name']").Attr("csk")
// 			awayTmFk := awayTmStr[0:3]
// 			awayTmNm := s.Find("[data-stat='away_team_name']").Text()

// 			gm := t.Game{
// 				Id:            fk,
// 				LeagueId:      "nfl",
// 				Fk:            fk,
// 				Url:           pbrGameRoot + fk,
// 				GameTime:      gmTime,
// 				GameResultsAt: gameRes,
// 				HomeTeamFk:    homeTmFk,
// 				HomeTeamName:  homeTmNm,
// 				AwayTeamFk:    awayTmFk,
// 				AwayTeamName:  awayTmNm,
// 			}
// 			fmt.Println(gm)
// 		}
// 	})

// 	return nil
// }
