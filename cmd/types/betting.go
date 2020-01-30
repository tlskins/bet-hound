package types

import (
	"fmt"
	"strings"
	"time"
)

// Changes

type BetRecipient struct {
	Id                *string `json:"id"`
	TwitterScreenName *string `json:"sn"`
}

type BetChanges struct {
	BetRecipient     BetRecipient       `json:"recipient"`
	EquationsChanges []*EquationChanges `json:"equationsChanges"`
	Delete           bool               `json:"delete"`
}

type EquationChanges struct {
	Id                int                        `json:"id"`
	Delete            *bool                      `json:"delete"`
	OperatorId        *int                       `json:"operatorId"`
	ExpressionChanges []*PlayerExpressionChanges `json:"expressionChanges"`
}

type PlayerExpressionChanges struct {
	Id       int     `json:"id"`
	IsLeft   *bool   `json:"is_left"`
	Delete   *bool   `json:"delete"`
	PlayerFk *string `json:"playerFk"`
	GameFk   *string `json:"gameFk"`
	MetricId *int    `json:"metricId"`
}

// Bet Maps

type BetMap struct {
	Id       int    `bson:"id" json:"id"`
	League   string `bson:"lg" json:"league"`
	Name     string `bson:"n" json:"name"`
	Field    string `bson:"f" json:"field"`
	LeftOnly bool   `bson:"lft" json:"left_only"`
	// FieldType  string `bson:"ft" json:"field_type"`
	// ResultType string `bson:"rt" json:"result_type"`
}

// Bet status

type BetStatus int

const (
	BetStatusPendingApproval BetStatus = iota
	BetStatusAccepted
	BetStatusFinal
	BetStatusExpired
	BetStatusDeclined
	BetStatusWithdrawn
)

func BetStatusFromString(s string) BetStatus {
	return map[string]BetStatus{
		"Pending Approval": BetStatusPendingApproval,
		"Accepted":         BetStatusAccepted,
		"Final":            BetStatusFinal,
		"Expired":          BetStatusExpired,
		"Declined":         BetStatusDeclined,
		"Withdrawn":        BetStatusWithdrawn,
	}[s]
}

func (s BetStatus) String() string {
	return map[BetStatus]string{
		BetStatusPendingApproval: "Pending Approval",
		BetStatusAccepted:        "Accepted",
		BetStatusFinal:           "Final",
		BetStatusExpired:         "Expired",
		BetStatusDeclined:        "Declined",
		BetStatusWithdrawn:       "Withdrawn",
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
	CreatedAt        *time.Time  `bson:"crt_at" json:"created_at"`
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
	if len(b.Recipient.UserName) > 0 {
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
			gmTime := expr.Game.GameTime
			// find earliest game start time that hasnt been played at the time of bet creation
			if minTime == nil || (gmTime.Before(*minTime) && gmTime.After(time.Now())) {
				minTime = &expr.Game.GameTime
			}
		}
	}

	return minTime
}

func (b Bet) maxFinalizedGameTime() *time.Time {
	var maxTime *time.Time
	for _, eq := range b.Equations {
		for _, expr := range eq.Expressions {
			gmTime := expr.Game.GameResultsAt
			// find latest game result time that hasnt been played at the time of bet creation
			if maxTime == nil || (gmTime.After(*maxTime) && gmTime.After(time.Now())) {
				maxTime = &expr.Game.GameResultsAt
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
	// if b.ExpiresAt.Before(time.Now()) {
	// 	errs = append(errs, "Invalid bet, all referenced games are already in progress or finalized.")
	// }
	// if b.FinalizedAt.Before(time.Now()) {
	// 	errs = append(errs, "Invalid bet, games are already final.")
	// }
	if len(b.Proposer.Id) == 0 {
		errs = append(errs, "No Proposer found.")
	}
	if len(b.Recipient.Id) == 0 {
		errs = append(errs, "No Proposer found.")
	}
	if b.Proposer.Id == b.Recipient.Id {
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

// Equation

type Equation struct {
	Id          int                 `bson:"id" json:"id"`
	Expressions []*PlayerExpression `bson:"exprs" json:"expressions"`
	Operator    *BetMap             `bson:"op" json:"operator"`
	Result      *bool               `bson:"res" json:"result"`
}

func (e Equation) LeftExpressions() (exprs []*PlayerExpression) {
	exprs = []*PlayerExpression{}
	for _, expr := range e.Expressions {
		if expr.IsLeft() {
			exprs = append(exprs, expr)
		}
	}
	return
}

func (e Equation) RightExpressions() (exprs []*PlayerExpression) {
	exprs = []*PlayerExpression{}
	for _, expr := range e.Expressions {
		if !expr.IsLeft() {
			exprs = append(exprs, expr)
		}
	}
	return
}

func (e Equation) Valid() error {
	if len(e.LeftExpressions()) == 0 {
		return fmt.Errorf("No left expressions found.")
	} else if len(e.RightExpressions()) == 0 {
		return fmt.Errorf("No right expressions found.")
	} else if e.Operator == nil {
		return fmt.Errorf("No operator found.")
	} else {
		for _, expr := range e.Expressions {
			if err := expr.Valid(); err != nil {
				return err
			}
		}
		return nil
	}
}

func (e Equation) String() (result string) {
	left, right := []string{}, []string{}
	for _, expr := range e.Expressions {
		if expr.Left {
			left = append(left, expr.String())
		} else {
			right = append(right, expr.String())
		}
	}
	return fmt.Sprintf(
		"%s %s %s",
		strings.Join(left, " "),
		e.Operator.Name,
		strings.Join(right, " "),
	)
}

func (e Equation) ResultString() string {
	if e.Result == nil {
		return e.String()
	}
	return fmt.Sprintf("%s (%t)", e.ResultString(), *e.Result)
}
