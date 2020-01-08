package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type BetMap struct {
	Name       string `bson:"n" json:"name"`
	Field      string `bson:"f" json:"field"`
	FieldType  string `bson:"ft" json:"field_type"`
	ResultType string `bson:"rt" json:"result_type"`
}

type BetCalc struct{}

func (b BetCalc) GreaterThan(v1, v2 float64) bool {
	return v1 > v2
}

func (b BetCalc) LesserThan(v1, v2 float64) bool {
	return v1 < v2
}

func (b BetCalc) EqualTo(v1, v2 float64) bool {
	return v1 == v2
}

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
	ExpiresAt        *time.Time  `bson:"exp_at" json:"expires_at"`
	FinalizedAt      *time.Time  `bson:"final_at" json:"finalized_at"`
	BetStatus        BetStatus   `bson:"status" json:"bet_status"`
	BetResult        *BetResult  `bson:"rslt" json:"result"`
}

func (b *Bet) AcceptBy(idStr, replyFk string) {
	if b.BetStatus.String() != "Pending Approval" {
		return
	}

	if b.ProposerReplyFk == nil && b.Proposer.IdStr == idStr {
		b.ProposerReplyFk = &replyFk
	} else if b.RecipientReplyFk == nil && b.Recipient.IdStr == idStr {
		b.RecipientReplyFk = &replyFk
	}

	if b.RecipientReplyFk != nil && b.ProposerReplyFk != nil {
		b.BetStatus = BetStatusAccepted
	}
}

func (b *Bet) CancelBy(idStr, replyFk string) {
	if b.BetStatus.String() != "Pending Approval" {
		return
	}

	if b.ProposerReplyFk == nil && b.Proposer.IdStr == idStr {
		b.ProposerReplyFk = &replyFk
		b.BetStatus = BetStatusCancelled
	} else if b.RecipientReplyFk == nil && b.Recipient.IdStr == idStr {
		b.RecipientReplyFk = &replyFk
		b.BetStatus = BetStatusCancelled
	}
}

func (b Bet) Response() (txt string) {
	if b.BetStatus.String() == "Pending Approval" {
		return fmt.Sprintf(
			"@%s @%s Is this correct: \"%s\" ? Reply \"Yes\"",
			b.Proposer.ScreenName,
			b.Recipient.ScreenName,
			b.String(),
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

func (b Bet) String() (result string) {
	result = fmt.Sprintf("%s bets", b.Proposer.Name)
	for _, eq := range b.Equations {
		result += fmt.Sprintf(" '%s'", eq.String())
	}
	return result
}

func (b Bet) minGameTime() *time.Time {
	var minTime *time.Time
	for _, eq := range b.Equations {
		allExprs := [][]*PlayerExpression{eq.LeftExpressions, eq.RightExpressions}
		for _, exprs := range allExprs {
			for _, expr := range exprs {
				// Toggle for expiration testing
				fmt.Println("min game time ", expr.Game.GameTime.String())
				if !expr.Game.Final && (minTime == nil || expr.Game.GameTime.Before(*minTime)) {
					// if minTime == nil || expr.Game.GameTime.Before(*minTime) {
					minTime = &expr.Game.GameTime
				}
			}
		}
	}

	return minTime
}

func (b Bet) maxFinalizedGameTime() *time.Time {
	var maxTime *time.Time
	for _, eq := range b.Equations {
		allExprs := [][]*PlayerExpression{eq.LeftExpressions, eq.RightExpressions}
		for _, exprs := range allExprs {
			for _, expr := range exprs {
				// Toggle for expiration testing
				if !expr.Game.Final && (maxTime == nil || expr.Game.GameResultsAt.After(*maxTime)) {
					// if maxTime == nil || expr.Game.GameTime.After(*maxTime) {
					maxTime = &expr.Game.GameResultsAt
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
	// Toggle for expiration testing
	if b.BetStatus.String() == "Pending Approval" && b.ExpiresAt != nil && time.Now().After(*b.ExpiresAt) {
		b.BetStatus = BetStatusFromString("Expired")
	}

	return nil
}

func (b Bet) Valid() error {
	errs := []string{}

	if len(b.Equations) == 0 {
		errs = append(errs, "Invalid bet syntax, no equations found.")
	}

	if b.ExpiresAt == nil {
		errs = append(errs, "Invalid bet, all referrenced games are already in progress or finalized.")
	}

	// if b.FinalizedAt == nil {
	// 	errs = append(errs, "Invalid bet, no final game time found.")
	// }

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

// Equation

type Equation struct {
	LeftExpressions  []*PlayerExpression `bson:"l_exprs" json:"left_expressions"`
	RightExpressions []*PlayerExpression `bson:"r_exprs" json:"right_expressions"`
	Action           *Word               `bson:"a_word" json:"action_word"`
	Operator         *BetMap             `bson:"op_word" json:"operator_word"`
	Delimiter        *Word               `bson:"delim_word" json:"delimiter_word"`
	Result           *bool               `bson:"res" json:"result"`
}

func (e Equation) Valid() error {
	if len(e.LeftExpressions) == 0 {
		return fmt.Errorf("No left expressions found.")
	} else if len(e.RightExpressions) == 0 {
		return fmt.Errorf("No right expressions found.")
		// } else if e.Action == nil {
		// 	return fmt.Errorf("No action found.")
	} else if e.Operator == nil {
		return fmt.Errorf("No operator found.")
		// } else if e.Delimiter == nil {
		// 	return fmt.Errorf("No delimiter found.")
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

func (e Equation) String() (result string) {
	left, right := "", ""
	for i, e := range e.LeftExpressions {
		if i > 0 {
			left += " "
		}
		left += e.String()
	}
	for i, e := range e.RightExpressions {
		if i > 0 {
			right += " "
		}
		right += e.String()
	}
	return fmt.Sprintf(
		"%s %s %s %s %s",
		left,
		e.Action.Text,
		e.Operator.Name,
		e.Delimiter.Text,
		right,
	)
}

func (e Equation) ResultDescription() (result string) {
	for i, e := range e.LeftExpressions {
		str := fmt.Sprintf(
			"%s. %s ",
			e.Player.FirstName[:1],
			e.Player.LastName,
		)
		if e.Value != nil {
			str += fmt.Sprintf("(%.2f)", *e.Value)
		} else {
			str += "(n/a)"
		}
		if i > 0 {
			str = " + " + str
		}
		result += str
	}

	if e.Operator.Name == ">" {
		result += " > "
	} else if e.Operator.Name == "<" {
		result += " < "
	}

	for i, e := range e.RightExpressions {
		str := fmt.Sprintf(
			"%s. %s ",
			e.Player.FirstName[:1],
			e.Player.LastName,
		)
		if e.Value != nil {
			str += fmt.Sprintf("(%.2f)", *e.Value)
		} else {
			str += "(n/a)"
		}
		if i > 0 {
			str = " + " + str
		}
		result += str
	}
	return result
}

// func (e Equation) expressionSources(left bool) (result string) {
// 	exprs := e.LeftExpressions
// 	if !left {
// 		exprs = e.RightExpressions
// 	}
// 	for i, expr := range exprs {
// 		if i > 0 {
// 			result += ", "
// 		}
// 		result += fmt.Sprintf(
// 			"%s. %s vs %s",
// 			expr.Player.FirstName[:1],
// 			expr.Player.LastName,
// 			expr.Game.VsTeamFk(expr.Player.TeamFk),
// 		)
// 	}
// 	return result
// }

// Expression

// Evaluates to a value. Metric and Player descendent of Action. Operator descendent of Metric.
type PlayerExpression struct {
	Player *Player  `bson:"player" json:"player"`
	Game   *Game    `bson:"gm" json:"game"`
	Value  *float64 `bson:"val" json:"value"`
	Metric *BetMap  `bson:"mtc" json:"metric"`
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

func (e PlayerExpression) String() (desc string) {
	vsTeam := e.Game.HomeTeamName
	if e.Player.TeamFk == e.Game.HomeTeamFk {
		vsTeam = e.Game.AwayTeamName
	}
	metric := ""
	if e.Metric != nil {
		metric = " " + e.Metric.Name
	}
	return fmt.Sprintf("%s.%s (%s-%s) vs %s%s",
		e.Player.FirstName[:1],
		e.Player.LastName,
		e.Player.TeamShort,
		e.Player.Position,
		vsTeam,
		metric,
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

func (m Metric) String() string {
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

func (m Metric) FixedValueMod() *float64 {
	for _, m := range m.Modifiers {
		f, err := strconv.ParseFloat(m.Text, 64)
		if err == nil {
			return &f
		}
	}
	return nil
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
	Game      *Game  `bson:"gm,omitempty" json:"game,omitempty"` // should only be aggregated
}
