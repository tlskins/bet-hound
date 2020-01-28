package types

import (
	"time"
)

type User struct {
	Id                string          `bson:"_id" json:"id"`
	Name              string          `bson:"nm" json:"name"`
	UserName          string          `bson:"usr_nm" json:"user_name"`
	Password          string          `bson:"pwd" json:"password"`
	Email             string          `bson:"em" json:"email"`
	TwitterUser       *TwitterUser    `bson:"twt" json:"twitter_user"`
	ViewedProfileLast *time.Time      `bson:"lstView" json:"viewed_profile_last"`
	Notifications     []*Notification `bson:"notes" json:"notificiations"`
	BetsWon           int             `bson:"bts_wn" json:"bets_won"`
	BetsLost          int             `bson:"bts_lst" json:"bets_lost"`
	InProgressBetIds  []string        `bson:"prg_bts" json:"in_progress_bet_ids"`
	PendingYouBetIds  []string        `bson:"pnd_u_bts" json:"pending_you_bet_ids"`
	PendingThemBetIds []string        `bson:"pnd_t_bts" json:"pending_them_bet_ids"`
}

type Notification struct {
	Id      string    `bson:"_id" json:"id"`
	SentAt  time.Time `bson:"snt_at" json:"sent_at"`
	Title   string    `bson:"ttl" json:"title"`
	Type    string    `bson:"typ" json:"type"`
	Message string    `bson:"msg" json:"message"`
}

type ProfileChanges struct {
	Name     *string `bson:"nm,omitempty" json:"name,omitempty"`
	UserName *string `bson:"usr_nm,omitempty" json:"user_name,omitempty"`
	Password *string `bson:"pwd,omitempty" json:"password,omitempty"`
}
