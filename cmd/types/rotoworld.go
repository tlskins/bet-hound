package types

import "time"

type RotoArticle struct {
	Id         string    `bson:"_id,omitempty" json:"id"`
	ImgSrc     string    `bson:"img" json:"img_src"`
	PlayerName string    `bson:"p_nm" json:"player_name"`
	Position   string    `bson:"pos" json:"position"`
	Team       string    `bson:"tm" json:"team"`
	Title      string    `bson:"ttl" json:"title"`
	Article    string    `bson:"art" json:"article"`
	ScrapedAt  time.Time `bson:"scp_at" json:"scraped_at"`
}

type RotoObserver struct {
	Observers []chan *RotoArticle
	Title     string
}
