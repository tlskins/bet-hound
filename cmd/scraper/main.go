package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func ExampleScrape() {
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
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#fantasy tr").Each(func(i int, s *goquery.Selection) {
		headTd := s.Find("td[data-stat=player]")
		name := headTd.Text()
		id, _ := headTd.Attr("data-append-csv")

		teamA := s.Find("td[data-stat=team] a")
		teamId := teamA.Text()
		teamName, _ := teamA.Attr("title")

		position := s.Find("td[data-stat=position]").Text()

		if len(id) > 0 {
			fmt.Printf("Player %d: %s %s %s %s %s\n", i, name, id, teamId, teamName, position)
		}
	})
}

func main() {
	ExampleScrape()
}
