package types

import (
	"fmt"
	// "strings"
	// "strconv"
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

// Bet result

type BetResult struct {
	Winner     User      `bson:"winner" json:"winner"`
	Loser      User      `bson:"loser" json:"loser"`
	Response   string    `bson:"resp" json:"response"`
	ResponseFk string    `bson:"resp_fk" json:"response_fk"`
	DecidedAt  time.Time `bson:"dec_at" json:"decided_at"`
}

// Bet

type Bet struct {
	Id               string      `bson:"_id" json:"id"`
	SourceFk         string      `bson:"source_fk" json:"source_fk"`
	Proposer         User        `bson:"proposer" json:"proposer"`
	Recipient        User        `bson:"recipient" json:"recipient"`
	AcceptFk         string      `bson:"acc_fk" json:"acc_fk"`
	ProposerReplyFk  *string     `bson:"pr_fk" json:"proposer_reply_fk"`
	RecipientReplyFk *string     `bson:"rr_fk" json:"recipient_reply_fk"`
	Equations        []*Equation `bson:"eqs" json:"equations"`
	ExpiresAt        time.Time   `bson:"exp_at" json:"expires_at"`
	FinalizedAt      time.Time   `bson:"final_at" json:"finalized_at"`
	BetStatus        BetStatus   `bson:"status" json:"bet_status"`
	BetResult        *BetResult  `bson:"rslt" json:"result"`
}

func (b Bet) Response() (txt string) {
	if b.BetStatus.String() == "Pending Approval" {
		return fmt.Sprintf(
			"@%s @%s Is this correct: \"%s\" ? Reply \"Yes\"",
			b.Proposer.ScreenName,
			b.Recipient.ScreenName,
			b.Description(),
		)
	} else if b.BetStatus.String() == "Accepted" {
		return fmt.Sprintf(
			"@%s @%s Bet recorded! When the bet has been finalized I will tweet the final results.",
			b.Proposer.ScreenName,
			b.Recipient.ScreenName,
		)
	} else if b.BetStatus.String() == "Final" {
		return b.BetResult.Response
	} else if b.BetStatus.String() == "Expired" {
		return fmt.Sprintf(
			"@%s @%s Bet has expired.",
			b.Proposer.ScreenName,
			b.Recipient.ScreenName,
		)
	} else if b.BetStatus.String() == "Cancelled" {
		return fmt.Sprintf(
			"@%s @%s Bet has been cancelled.",
			b.Proposer.ScreenName,
			b.Recipient.ScreenName,
		)
	}
	return ""
}

func (b Bet) Description() (result string) {
	result = fmt.Sprintf("%s bets", b.Proposer.Name)
	for _, eq := range b.Equations {
		result += fmt.Sprintf(" '%s'", eq.Description())
	}
	return result
}

// Equation

type Equation struct {
	LeftExpressions  []*PlayerExpression `bson:"l_expr" json:"left_expressions"`
	RightExpressions []*PlayerExpression `bson:"l_expr" json:"left_expressions"`
	Metric           *Metric             `bson:"metric" json:"metric"`
	Action           *Word               `bson:"a_word" json:"action_word"`
	Operator         *Word               `bson:"op_word" json:"operator_word"`
	Delimiter        *Word               `bson:"delim_word" json:"delimiter_word"`
	Result           *bool               `bson:"res" json:"result"`
}

func (e Equation) Valid() error {
	if len(e.LeftExpressions) == 0 {
		return fmt.Errorf("No left expressions found.")
	} else if len(e.RightExpressions) == 0 {
		return fmt.Errorf("No right expressions found.")
	} else if e.Metric == nil {
		return fmt.Errorf("No metric found.")
	} else if e.Metric.Valid() != nil {
		return e.Metric.Valid()
	} else if e.Action == nil {
		return fmt.Errorf("No action found.")
	} else if e.Operator == nil {
		return fmt.Errorf("No operator found.")
	} else if e.Delimiter == nil {
		return fmt.Errorf("No delimiter found.")
	} else {
		for _, l := range e.LeftExpressions {
			err := l.Valid()
			if err != nil {
				return err
			}
		}
		for _, r := range e.RightExpressions {
			err := r.Valid()
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (e Equation) Description() (result string) {
	return fmt.Sprintf(
		"%s %s %s %s %s %s",
		e.expressionSources(true),
		e.Action.Text,
		e.Operator.Text,
		e.Metric.Description(),
		e.Delimiter.Text,
		e.expressionSources(false),
	)
}

func (e Equation) ResultDescription() (result string) {
	for i, e := range e.LeftExpressions {
		str := "n/a"
		if e.Value != nil {
			str += fmt.Sprintf(
				"%s. %s (%f)",
				e.Player.FirstName[:1],
				e.Player.LastName,
				*e.Value,
			)
		}
		if i > 0 {
			str = " + " + str
		}
		result += str
	}

	if e.Operator.Lemma == "more" {
		result += " > "
	} else if e.Operator.Lemma == "less" || e.Operator.Lemma == "few" {
		result += " < "
	}

	for i, e := range e.RightExpressions {
		str := "n/a"
		if e.Value != nil {
			str += fmt.Sprintf(
				"%s. %s (%f)",
				e.Player.FirstName[:1],
				e.Player.LastName,
				*e.Value,
			)
		}
		if i > 0 {
			str = " + " + str
		}
		result += str
	}
	return result
}

func (e Equation) expressionSources(left bool) (result string) {
	exprs := e.LeftExpressions
	if !left {
		exprs = e.RightExpressions
	}
	for i, expr := range exprs {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf(
			"%s. %s vs %s",
			expr.Player.FirstName[:1],
			expr.Player.LastName,
			expr.Game.VsTeamFk(expr.Player.TeamFk),
		)
	}
	return result
}

// Expression

// Evaluates to a value. Metric and Player descendent of Action. Operator descendent of Metric.
type PlayerExpression struct {
	Player *Player  `bson:"player" json:"player"`
	Game   *Game    `bson:"gm" json:"game"`
	Value  *float64 `bson:"val" json:"value"`
}

func (e PlayerExpression) Valid() error {
	if e.Player == nil {
		return fmt.Errorf("Player not found.")
	} else if e.Game == nil {
		return fmt.Errorf("Game not found for player %s.", e.Player.Name)
	} else {
		return nil
	}
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

// Metric

type Metric struct {
	Word      *Word   `bson:"word" json:"word"`
	Modifiers []*Word `bson:"mods" json:"modifiers"`
}

func (m Metric) Valid() error {
	if m.Word == nil {
		return fmt.Errorf("Metric not found.")
	} else {
		return nil
	}
}

func (m Metric) Description() string {
	result := m.Word.Text
	for _, m := range m.Modifiers {
		result += " " + m.Text
	}
	return result
}

func (m Metric) PPR() float64 {
	for _, m := range m.Modifiers {
		if m.Text == "ppr" {
			return 1.0
		} else if m.Text == "0.5ppr" || m.Text == ".5ppr" {
			return 0.5
		}
	}
	return 0.0
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
