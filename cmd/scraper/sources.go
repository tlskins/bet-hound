package scraper

import (
	"fmt"
	gq "github.com/PuerkitoBio/goquery"
	"log"
	"net/http"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

func ScrapeSources() {
	// Request the HTML page.
	res, err := http.Get("https://www.pro-football-reference.com/years/2019/fantasy.htm")
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

	var sources []*t.Source

	doc.Find("#fantasy tr").Each(func(i int, s *gq.Selection) {
		headTd := s.Find("td[data-stat=player]")
		name := headTd.Text()
		id, _ := headTd.Attr("data-append-csv")
		url, _ := headTd.Find("a").Attr("href")

		teamA := s.Find("td[data-stat=team] a")
		teamId := teamA.Text()
		teamName, _ := teamA.Attr("title")

		position := s.Find("td[data-stat=position]").Text()

		if len(id) > 0 {
			fmt.Printf("Player %d: %s %s %s %s %s %s\n", i, name, id, teamId, teamName, position, url)
			sources = append(sources, &t.Source{
				Name:     &name,
				Fk:       &id,
				TeamFk:   &teamId,
				TeamName: &teamName,
				Position: &position,
				Url:      &url,
			})
		}
	})

	for _, source := range sources {
		fmt.Println("source: ", *source)
	}
	err = db.UpsertSources(&sources)
	if err != nil {
		log.Fatal(err)
	}
}