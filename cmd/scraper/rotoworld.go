package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	t "bet-hound/cmd/types"

	gq "github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/satori/go.uuid"
)

func getRotoNflHtml() (html string, err error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// run task list
	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://www.rotoworld.com/football/nfl/player-news`),
		// chromedp.WaitVisible(`.player-news-article`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	)

	return html, err
}

func RotoNflArticles(numResults int) (articles []*t.RotoArticle, err error) {
	fmt.Println("scraping rotoworld nfl")
	html, err := getRotoNflHtml()
	if err != nil {
		return
	}
	doc, err := gq.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return
	}

	articles = make([]*t.RotoArticle, numResults)

	doc.Find("#player-news-page-wrapper > div > div > div.player-news.default > ul > li").Each(func(i int, s *gq.Selection) {
		if i >= numResults {
			return
		}

		imgSrc, _ := s.Find(".player-news-article .player-news-article__header .player-news-article__logo").Attr("src")
		name := s.Find(".player-news-article .player-news-article__header .player-news-article__profile__name a").Text()
		posRaw := s.Find(".player-news-article .player-news-article__header .player-news-article__profile__position").Text()
		re := regexp.MustCompile(`\w+`)
		pos := string(re.Find([]byte(posRaw)))
		team := s.Find(".player-news-article .player-news-article__header .player-news-article__profile__position a").Text()
		title := s.Find(".player-news-article .player-news-article__body .player-news-article__title h3").Text()
		article := s.Find(".player-news-article .player-news-article__body .player-news-article__summary p").Text()

		articles[i] = &t.RotoArticle{
			Id:         uuid.NewV4().String(),
			PlayerName: name,
			ImgSrc:     imgSrc,
			Position:   pos,
			Team:       team,
			Title:      title,
			Article:    article,
			ScrapedAt:  time.Now(),
		}
	})

	return articles, nil
}
