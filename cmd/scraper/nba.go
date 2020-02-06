package scraper

import (
	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	gq "github.com/PuerkitoBio/goquery"
)

var nbaLgId = "nba"
var pbrTz = "-0500 EST"
var pbrLoc = "America/New_York"
var pbrUri = "https://www.basketball-reference.com"
var pbrTeamsUrl = "https://www.basketball-reference.com/teams/"
var pbrSchedsRoot = "https://www.basketball-reference.com/leagues/NBA_2020_games-%s.htm"
var pbrGameRoot = "https://www.basketball-reference.com/boxscores/%s.html"
var pbrPlayerFkRgx *regexp.Regexp = regexp.MustCompile(".+\\/(.+).html")
var pbrLocRgx *regexp.Regexp = regexp.MustCompile("Location:[\\s]+([a-zA-Z ]+),")
var pbrTeamFkRgx *regexp.Regexp = regexp.MustCompile("\\/teams\\/(.+)\\/")

func ScrapeNbaTeams() {
	fmt.Printf("%s: Scraping nba teams...\n", time.Now().String())
	teamUrls := scrapeNbaTeamUrls()
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

func ScrapeNbaPlayers() {
	fmt.Printf("%s: Scraping nba players...\n", time.Now().String())
	teamUrls := scrapeNbaTeamUrls()
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
				name := fmt.Sprintf("%s at %s (%s)", awayTmNm, homeTmNm, gmTime.Format("Jan 2 2006"))
				gm := t.Game{
					Id:            nbaLgId + fk,
					LeagueId:      nbaLgId,
					Fk:            fk,
					Name:          name,
					Url:           fmt.Sprintf(pbrGameRoot, fk),
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

func ScrapeNbaGameLog(game *t.GameAndLog) {
	doc, err := GetGqDocument(game.Url)
	if err != nil {
		panic(err)
	}

	gameLog := &t.GameLog{}
	awayScoreTxt := doc.Find("#content > div.scorebox > div:nth-child(1) > div.scores").Text()
	awayScoreTxt = strings.TrimSpace(awayScoreTxt)
	awayScore, awayScoreErr := strconv.ParseFloat(awayScoreTxt, 64)
	awayTeamNm := doc.Find("#content > div.scorebox > div:nth-child(1) > div:nth-child(1) > strong > a").Text()
	homeScoreTxt := doc.Find("#content > div.scorebox > div:nth-child(2) > div.scores").Text()
	homeScoreTxt = strings.TrimSpace(homeScoreTxt)
	homeScore, homeScoreErr := strconv.ParseFloat(homeScoreTxt, 64)
	homeTeamNm := doc.Find("#content > div.scorebox > div:nth-child(2) > div:nth-child(1) > strong > a").Text()

	if awayScoreErr != nil {
		panic(awayScoreErr)
	}
	if homeScoreErr != nil {
		panic(homeScoreErr)
	}

	gameLog.AwayTeamLog = t.TeamLog{
		Fk:       game.AwayTeamFk,
		TeamName: awayTeamNm,
		Score:    awayScore,
	}
	gameLog.HomeTeamLog = t.TeamLog{
		Fk:       game.HomeTeamFk,
		TeamName: homeTeamNm,
		Score:    homeScore,
	}
	gameLog.EvaluateWinner()
	gameLog.PlayerLogs = scrapeNbaPlayerLogs(doc)
	game.GameLog = gameLog
	games := []*t.GameAndLog{game}

	db.UpsertGameAndLogs(&games)
}

// helpers

func scrapeNbaPlayerLogs(doc *gq.Document) (playerLogs map[string]*t.PlayerLog) {
	playerLogs = make(map[string]*t.PlayerLog)
	doc.Find("#box-SAS-game-basic > tbody > tr").Each(func(i int, s *gq.Selection) {
		playerFk, _ := s.Find("[data-stat='player']").Attr("data-append-csv")
		played := s.Find("[data-stat='mp']").Text()
		if len(played) > 0 && len(playerFk) > 0 {
			log := t.NbaPlayerLog{}
			s.Find("td").Each(func(i int, s *gq.Selection) {
				data, _ := s.Attr("data-stat")
				switch data {
				case "mp":
					log.MinsPlayed, _ = strconv.ParseFloat(s.Text(), 64)
				case "fg":
					log.FieldGoals, _ = strconv.ParseFloat(s.Text(), 64)
				case "fga":
					log.FieldGoalAtts, _ = strconv.ParseFloat(s.Text(), 64)
				case "fg_pct":
					log.FieldGoalPct, _ = strconv.ParseFloat(s.Text(), 64)
				case "fg3":
					log.FieldGoal3s, _ = strconv.ParseFloat(s.Text(), 64)
				case "fg3a":
					log.FieldGoal3Atts, _ = strconv.ParseFloat(s.Text(), 64)
				case "fg3_pct":
					log.FieldGoal3Pct, _ = strconv.ParseFloat(s.Text(), 64)
				case "ft":
					log.FreeThrows, _ = strconv.ParseFloat(s.Text(), 64)
				case "fta":
					log.FreeThrowAtts, _ = strconv.ParseFloat(s.Text(), 64)
				case "ft_pct":
					log.FreeThrowPct, _ = strconv.ParseFloat(s.Text(), 64)
				case "orb":
					log.OffRebound, _ = strconv.ParseFloat(s.Text(), 64)
				case "drb":
					log.DefRebound, _ = strconv.ParseFloat(s.Text(), 64)
				case "trb":
					log.TotalRebounds, _ = strconv.ParseFloat(s.Text(), 64)
				case "ast":
					log.Assists, _ = strconv.ParseFloat(s.Text(), 64)
				case "stl":
					log.Steals, _ = strconv.ParseFloat(s.Text(), 64)
				case "blk":
					log.Blocks, _ = strconv.ParseFloat(s.Text(), 64)
				case "tov":
					log.TurnOvers, _ = strconv.ParseFloat(s.Text(), 64)
				case "pf":
					log.PersonalFouls, _ = strconv.ParseFloat(s.Text(), 64)
				case "pts":
					log.Points, _ = strconv.ParseFloat(s.Text(), 64)
				case "plus_minus":
					log.PlusMinus, _ = strconv.ParseFloat(s.Text(), 64)
				}
			})
			var playerLog t.PlayerLog = log
			playerLogs[playerFk] = &playerLog
		}
	})
	return
}

func scrapeNbaTeamUrls() (teamUrls []string) {
	fmt.Printf("%s: Scraping nba team FKs...\n", time.Now().String())
	doc, err := GetGqDocument(pbrTeamsUrl)
	if err != nil {
		panic(err)
	}

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

func nbaPlayerFkFrom(text string) (fk string) {
	fkMatch := pbrPlayerFkRgx.FindStringSubmatch(text)
	if len(fkMatch) > 1 {
		fk = fkMatch[1]
		fk = strings.ToLower(fk)
	}
	return
}
