package types

import (
	"fmt"
	// "strings"
	// "time"
)

// Bet status

type BetStatus int

const (
	BetStatusPendingProposer BetStatus = iota
	BetStatusPendingRecipient
	BetStatusAccepted
	BetStatusFinal
)

func BetStatusFromString(s string) BetStatus {
	return map[string]BetStatus{
		"Pending Proposer":  BetStatusPendingProposer,
		"Pending Recipient": BetStatusPendingRecipient,
		"Accepted":          BetStatusAccepted,
		"Final":             BetStatusFinal,
	}[s]
}

func (s BetStatus) String() string {
	return map[BetStatus]string{
		BetStatusPendingProposer:  "Pending Proposer",
		BetStatusPendingRecipient: "Pending Recipient",
		BetStatusAccepted:         "Accepted",
		BetStatusFinal:            "Final",
	}[s]
}

// Bet

type Bet struct {
	Id               string     `bson:"_id" json:"id"`
	SourceFk         int64      `bson:"source_fk" json:"source_fk"`
	Proposer         User       `bson:"proposer" json:"proposer"`
	Recipient        User       `bson:"recipient" json:"recipient"`
	BetStatus        BetStatus  `bson:"status" json:"bet_status"`
	ProposerCheckFk  int64      `bson:"p_chk_fk" json:"proposer_check_fk"`
	RecipientCheckFk int64      `bson:"r_chk_fk" json:"recipient_check_fk"`
	Equations        []Equation `bson:"eqs" "json:"equations"`
}

func (b Bet) Response() (txt string) {
	return "test"
	// if b.BetStatus.String() == "Pending Proposer" {
	// 	return fmt.Sprintf("%s%s Is this correct: \"%s\" ? Reply \"Yes\"", "@", b.Proposer.ScreenName, b.Text())
	// } else if b.BetStatus.String() == "Pending Recipient" {
	// 	return fmt.Sprintf("%s%s Do you accept this bet? : \"%s\" Reply \"Yes\"", "@", b.Recipient.ScreenName, b.Text())
	// }
	// return fmt.Sprintf("%s%s Bet recorded! When the bet has been finalized I will tweet the final results", b.Proposer.ScreenName, b.Recipient.ScreenName)
}

// Equation

type Equation struct {
	// LeftExpression  Expression     `bson:"l_exp" json:"left_expression"`
	// RightExpression Expression     `bson:"r_exp" json:"right_expression"`
	LeftExpression  PlayerExpression `bson:"l_exp" json:"left_expression"`
	RightExpression PlayerExpression `bson:"r_exp" json:"right_expression"`
	Operator        OperatorPhrase   `bson:"m_phrase" json:"metric_phrase"`
	// TODO : Add complete function to check event time / metric exists
}

func (e Equation) Text() (txt string) {
	return fmt.Sprintf("%s %s %s",
		e.LeftExpression.Description(),
		e.Operator.Text(),
		e.RightExpression.Description(),
	)
}

// Operator Phrase

type OperatorPhrase struct {
	ActionWord   Word       `bson:"a_word" json:"action_word"`
	OperatorWord Word       `bson:"op_word" json:"operator_word"`
	Metric       *Metric    `bson:"metric" json:"metric"`
	EventTime    *EventTime `bson:"event_time" json:"event_time"`
}

func (p OperatorPhrase) Text() string {
	return p.ActionWord.Text + " " + p.OperatorWord.Text
}

// Metric / Event Time

type Metric struct {
	Word      Word     `bson:"word" json:"word"`
	Modifiers []string `bson:"mods" json:"modifiers"`
}

type EventTime struct {
	Text      string   `bson:"txt" json:"text"`
	Lemma     string   `bson:"lemma" json:"lemma'`
	Modifiers []string `bson:"mods" json:"modifiers"`
}

// Expression

type Expression interface {
	Description() string
	ShortDescription() string
	Value() *float64
	// Add complete? function to check if game / metric exist
}

type PlayerExpression struct {
	Player    Player     `bson:"player" json:"player"`
	Game      *Game      `bson:"gm" json:"game"`
	Metric    *Metric    `bson:"metric" json:"metric"`         // TODO : if metric not exists point to operatoer phrase metric
	EventTime *EventTime `bson:"event_time" json:"event_time"` // TODO : if event time not exists point to operatoer event time
}

func (e PlayerExpression) Description() (desc string) {
	vsTeam := e.Game.HomeTeamName
	if e.Player.TeamFk == e.Game.HomeTeamFk {
		vsTeam = e.Game.AwayTeamName
	}
	return fmt.Sprintf("%s.%s (%s-%s) vs %s",
		e.Player.FirstName[:1],
		e.Player.LastName,
		e.Player.TeamShort,
		e.Player.Position,
		vsTeam,
	)
}

func (e PlayerExpression) ShortDescription() (desc string) {
	return fmt.Sprintf("%s.%s",
		e.Player.FirstName[:1],
		e.Player.LastName,
	)
}

func (e PlayerExpression) Value() (value float64) {
	return 1.0
}

// Player

type Player struct {
	Id        string `bson:"_id" json:"id"`
	Name      string `bson:"name,omitempty" json:"name"`
	FirstName string `bson:"f_name,omitempty" json:"first_name"`
	LastName  string `bson:"l_name,omitempty" json:"last_name"`
	Fk        string `bson:"fk,omitempty" json:"fk"`
	TeamFk    string `bson:"team_fk,omitempty" json:"team_fk"`
	TeamName  string `bson:"team_name,omitempty" json:"team_name"`
	TeamShort string `bson:"team_short,omitempty" json:"team_short"`
	Position  string `bson:"pos,omitempty" json:"position"`
	Url       string `bson:"url,omitempty" json:"url"`
}
