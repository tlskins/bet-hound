package types

import (
	"fmt"
	"strings"
	"time"
)

// Bet

type Bet struct {
	Id               string      `bson:"_id" json:"id"`
	LeagueId         string      `bson:"lg_id" json:"league_id"`
	CreatedAt        *time.Time  `bson:"crt_at" json:"created_at"`
	SourceFk         string      `bson:"source_fk" json:"source_fk"`
	Proposer         IndexUser   `bson:"proposer" json:"proposer"`
	Recipient        *IndexUser  `bson:"recipient" json:"recipient"`
	AcceptFk         string      `bson:"acc_fk" json:"acc_fk"`
	ProposerReplyFk  *string     `bson:"pr_fk" json:"proposer_reply_fk"`
	RecipientReplyFk *string     `bson:"rr_fk" json:"recipient_reply_fk"`
	Equations        []*Equation `bson:"eqs" json:"equations"`
	ExpiresAt        *time.Time  `bson:"exp_at" json:"expires_at"`
	FinalizedAt      *time.Time  `bson:"final_at" json:"finalized_at"`
	BetStatus        BetStatus   `bson:"status" json:"bet_status"`
	BetResult        *BetResult  `bson:"rslt" json:"result"`
}

func (b Bet) ProposerName() string {
	if len(b.Proposer.UserName) > 0 {
		return b.Proposer.UserName
	} else if b.Proposer.TwitterUser != nil {
		return b.Proposer.TwitterUser.ScreenName
	} else {
		return "?"
	}
}

func (b Bet) RecipientName() string {
	if b.Recipient == nil {
		return "*Anyone"
	} else if len(b.Recipient.UserName) > 0 {
		return b.Recipient.UserName
	} else if b.Recipient.TwitterUser != nil {
		return b.Recipient.TwitterUser.ScreenName
	} else {
		return "?"
	}
}

func (b Bet) TwitterHandles() (result string) {
	handles := []string{}
	if b.Proposer.TwitterUser != nil {
		handles = append(handles, "@"+b.Proposer.TwitterUser.ScreenName)
	}
	if b.Recipient.TwitterUser != nil {
		handles = append(handles, "@"+b.Recipient.TwitterUser.ScreenName)
	}
	return strings.Join(handles, " ")
}

func (b Bet) String() (result string) {
	result = fmt.Sprintf("%s bets", b.Proposer.Name)
	for _, eq := range b.Equations {
		result += fmt.Sprintf(" '%s'", eq.String())
	}
	return result
}

func (b Bet) ResultString() string {
	results := []string{}
	for _, eq := range b.Equations {
		results = append(results, eq.ResultString())
	}
	return strings.Join(results, "\n")
}

func (b Bet) minGameTime() *time.Time {
	var minTime *time.Time
	for _, eq := range b.Equations {
		for _, expr := range eq.Expressions {
			gm := expr.GetGame()
			if gm != nil {
				// find earliest game start time that hasnt been played at the time of bet creation
				if minTime == nil || (gm.GameTime.Before(*minTime) && gm.GameTime.After(time.Now())) {
					minTime = &gm.GameTime
				}
			}
		}
	}

	return minTime
}

func (b Bet) maxFinalizedGameTime() *time.Time {
	var maxTime *time.Time
	for _, eq := range b.Equations {
		for _, expr := range eq.Expressions {
			gm := expr.GetGame()
			if gm != nil {
				if maxTime == nil || (gm.GameResultsAt.After(*maxTime)) {
					maxTime = &gm.GameResultsAt
				}
			}
		}
	}

	return maxTime
}

func (b *Bet) PostProcess() error {
	if b.ExpiresAt == nil {
		b.ExpiresAt = b.minGameTime()
	}
	if b.FinalizedAt == nil {
		b.FinalizedAt = b.maxFinalizedGameTime()
	}

	return nil
}

func (b Bet) Valid() error {
	errs := []string{}
	if len(b.Equations) == 0 {
		errs = append(errs, "No bet details found.")
	}
	if b.ExpiresAt.Before(time.Now()) {
		errs = append(errs, "Invalid bet, all referenced games are already in progress or finalized.")
	}
	if b.FinalizedAt.Before(time.Now()) {
		errs = append(errs, "Invalid bet, games are already final.")
	}
	if len(b.Proposer.Id) == 0 {
		errs = append(errs, "No Proposer found.")
	}
	if b.Recipient != nil && b.Proposer.Id == b.Recipient.Id {
		errs = append(errs, "Proposer cant be the same as the recipient.")
	}
	for _, eq := range b.Equations {
		err := eq.Valid()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, " "))
	} else {
		return nil
	}
}

// for unmarshalling then converting to bet

type MongoBet struct {
	Id               string           `bson:"_id" json:"id"`
	LeagueId         string           `bson:"lg_id" json:"league_id"`
	CreatedAt        *time.Time       `bson:"crt_at" json:"created_at"`
	SourceFk         string           `bson:"source_fk" json:"source_fk"`
	Proposer         IndexUser        `bson:"proposer" json:"proposer"`
	Recipient        *IndexUser       `bson:"recipient" json:"recipient"`
	AcceptFk         string           `bson:"acc_fk" json:"acc_fk"`
	ProposerReplyFk  *string          `bson:"pr_fk" json:"proposer_reply_fk"`
	RecipientReplyFk *string          `bson:"rr_fk" json:"recipient_reply_fk"`
	Equations        []*MongoEquation `bson:"eqs" json:"equations"`
	ExpiresAt        *time.Time       `bson:"exp_at" json:"expires_at"`
	FinalizedAt      *time.Time       `bson:"final_at" json:"finalized_at"`
	BetStatus        BetStatus        `bson:"status" json:"bet_status"`
	BetResult        *BetResult       `bson:"rslt" json:"result"`
}

func (m MongoBet) Bet() *Bet {
	eqs := make([]*Equation, len(m.Equations))
	for i, eq := range m.Equations {
		eqs[i] = eq.Equation()
	}

	return &Bet{
		Id:               m.Id,
		LeagueId:         m.LeagueId,
		CreatedAt:        m.CreatedAt,
		SourceFk:         m.SourceFk,
		Proposer:         m.Proposer,
		Recipient:        m.Recipient,
		AcceptFk:         m.AcceptFk,
		ProposerReplyFk:  m.ProposerReplyFk,
		RecipientReplyFk: m.RecipientReplyFk,
		ExpiresAt:        m.ExpiresAt,
		FinalizedAt:      m.FinalizedAt,
		BetStatus:        m.BetStatus,
		BetResult:        m.BetResult,
		Equations:        eqs,
	}
}

// grouped bets

type BetsResponse struct {
	AcceptedBets      []*Bet `json:"accepted_bets"`
	FinalBets         []*Bet `json:"final_bets"`
	PublicPendingBets []*Bet `json:"public_pending_bets"`
	PendingBets       []*Bet `json:"pending_bets"`
	ClosedBets        []*Bet `json:"closed_bets"`
}
