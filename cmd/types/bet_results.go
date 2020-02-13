package types

import "time"

// Bet result

type BetResult struct {
	Winner     IndexUser `bson:"winner" json:"winner"`
	Loser      IndexUser `bson:"loser" json:"loser"`
	Response   string    `bson:"resp" json:"response"`
	ResponseFk string    `bson:"resp_fk" json:"response_fk"`
	DecidedAt  time.Time `bson:"dec_at" json:"decided_at"`
}
