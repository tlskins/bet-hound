package types

import (
	"fmt"
	// "strings"
	"time"
)

// Bet status

type BetStatus int

const (
	BetStatusPendingApproval BetStatus = iota
	BetStatusAccepted
	BetStatusFinal
	BetStatusExpired
	BetStatusCancelled
)

func BetStatusFromString(s string) BetStatus {
	return map[string]BetStatus{
		"Pending Approval": BetStatusPendingApproval,
		"Accepted":         BetStatusAccepted,
		"Final":            BetStatusFinal,
		"Expired":          BetStatusExpired,
		"Cancelled":        BetStatusCancelled,
	}[s]
}

func (s BetStatus) String() string {
	return map[BetStatus]string{
		BetStatusPendingApproval: "Pending Approval",
		BetStatusAccepted:        "Accepted",
		BetStatusFinal:           "Final",
		BetStatusExpired:         "Expired",
		BetStatusCancelled:       "Cancelled",
	}[s]
}

// Bet

type Bet struct {
	Id               string    `bson:"_id" json:"id"`
	SourceFk         string    `bson:"source_fk" json:"source_fk"`
	Proposer         User      `bson:"proposer" json:"proposer"`
	Recipient        User      `bson:"recipient" json:"recipient"`
	BetStatus        BetStatus `bson:"status" json:"bet_status"`
	AcceptFk         string    `bson:"acc_fk" json:"acc_fk"`
	ProposerReplyFk  *string   `bson:"pr_fk" json:"proposer_reply_fk"`
	RecipientReplyFk *string   `bson:"rr_fk" json:"recipient_reply_fk"`
	Equation         Equation  `bson:"eq" "json:"equation"`
	Result           string    `bson:"result" json:"result"`
	FinalizedAt      time.Time `bson:"final_at" json:"finalized_at"`
}

func (b Bet) Response() (txt string) {
	if b.BetStatus.String() == "Pending Approval" {
		return fmt.Sprintf(
			"@%s @%s Is this correct: \"%s\" ? Reply \"Yes\"",
			b.Proposer.ScreenName,
			b.Recipient.ScreenName,
			b.Text(),
		)
	} else if b.BetStatus.String() == "Accepted" {
		return fmt.Sprintf(
			"@%s @%s Bet recorded! When the bet has been finalized I will tweet the final results.",
			b.Proposer.ScreenName,
			b.Recipient.ScreenName,
		)
	} else if b.BetStatus.String() == "Final" {
		return b.Result
	} else if b.BetStatus.String() == "Expired" {
		return "Bet has expired."
	} else if b.BetStatus.String() == "Cancelled" {
		return "Bet has been cancelled."
	}
	return ""
}

func (b Bet) Text() (txt string) {
	eq := b.Equation
	metric, rightMetric := eq.MetricString()
	if len(rightMetric) > 0 {
		rightMetric = " " + eq.Operator.ActionWord.Lemma + rightMetric
	}
	return fmt.Sprintf("%s bets %s %s %s %s %s%s",
		b.Proposer.Name,
		eq.LeftExpression.Description(),
		eq.Operator.Text(),
		metric,
		"than",
		eq.RightExpression.Description(),
		rightMetric,
	)
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

func (e Equation) Complete() (err error) {
	err = e.Operator.Complete()
	if err != nil {
		return err
	}
	err, _ = e.LeftExpression.Complete()
	if err != nil {
		return err
	}
	err, _ = e.RightExpression.Complete()
	if err != nil {
		return err
	}
	// if rFinal && lFinal {
	// 	return fmt.Errorf("Both games are already final!")
	// }
	return nil
}

func (e Equation) Metrics() (lftMetric Metric, rgtMetric Metric) {
	if e.LeftExpression.Metric != nil {
		lftMetric = *e.LeftExpression.Metric
		if e.RightExpression.Metric == nil {
			rgtMetric = *e.LeftExpression.Metric
		}
	} else {
		lftMetric = *e.RightExpression.Metric
		rgtMetric = *e.RightExpression.Metric
	}

	return lftMetric, rgtMetric
}

func (e Equation) MetricString() (metricStr string, rightMetric string) {
	metric := e.LeftExpression.Metric
	if metric == nil {
		metric = e.RightExpression.Metric
	}
	for _, m := range metric.Modifiers {
		metricStr = m + " " + metricStr
	}
	metricStr = metricStr + metric.Word.Lemma + "s"
	if e.RightExpression.Metric != nil {
		rightMetric = e.RightExpression.Metric.Word.Lemma
		for _, m := range metric.Modifiers {
			rightMetric = metricStr + " " + m
		}
	}
	return metricStr, rightMetric
}

// Operator Phrase

type OperatorPhrase struct {
	ActionWord   Word `bson:"a_word" json:"action_word"`
	OperatorWord Word `bson:"op_word" json:"operator_word"`
}

func (p OperatorPhrase) Complete() (err error) {
	if p.ActionWord.Lemma == "" && p.OperatorWord.Lemma == "" {
		return fmt.Errorf("Invalid bet syntax.")
	} else {
		return nil
	}
}

func (p OperatorPhrase) Text() string {
	return p.ActionWord.Text + " " + p.OperatorWord.Text
}

// Metric / Event Time

type Metric struct {
	Word      Word     `bson:"word" json:"word"`
	Modifiers []string `bson:"mods" json:"modifiers"`
}

func (m Metric) PPR() float64 {
	for _, m := range m.Modifiers {
		if m == "ppr" {
			return 1.0
		} else if m == "0.5ppr" || m == ".5ppr" {
			return 0.5
		}
	}
	return 0.0
}

type EventTime struct {
	Word      Word     `bson:"word" json:"word"`
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

func (e PlayerExpression) Complete() (err error, final bool) {
	if len(e.Player.Id) == 0 {
		return fmt.Errorf("Player not found."), e.Game.Final
	} else if e.Game == nil {
		return fmt.Errorf("Game not found."), e.Game.Final
	}
	return nil, e.Game.Final
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
	return fmt.Sprintf("%s.%s %s %s",
		e.Player.FirstName[:1],
		e.Player.LastName,
		"vs",
		e.Game.VsTeamFk(e.Player.TeamFk),
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
