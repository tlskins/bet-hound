package types

import (
	"fmt"
	"strings"
	"time"
)

// Changes

type BetChanges struct {
	RecipientId      string             `json:"recipientId"`
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
	Id         int    `bson:"id" json:"id"`
	League     string `bson:"lg" json:"league"`
	Name       string `bson:"n" json:"name"`
	Field      string `bson:"f" json:"field"`
	FieldType  string `bson:"ft" json:"field_type"`
	ResultType string `bson:"rt" json:"result_type"`
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
			if minTime == nil || expr.Game.GameTime.Before(*minTime) {
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
			if maxTime == nil || expr.Game.GameResultsAt.After(*maxTime) {
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
		if expr.IsLeft {
			exprs = append(exprs, expr)
		}
	}
	return
}

func (e Equation) RightExpressions() (exprs []*PlayerExpression) {
	exprs = []*PlayerExpression{}
	for _, expr := range e.Expressions {
		if !expr.IsLeft {
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
		if expr.IsLeft {
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

// Expression

// Evaluates to a value. Metric and Player descendent of Action. Operator descendent of Metric.
type PlayerExpression struct {
	Id     int      `bson:"id" json:"id"`
	IsLeft bool     `bson:"lft" json:"is_left"`
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
	} else if e.Metric == nil {
		return fmt.Errorf("Metric not found for player %s.", e.Player.Name)
	} else {
		return nil
	}
}

func (e PlayerExpression) String() (desc string) {
	if e.Player == nil || e.Game == nil {
		return "?"
	}
	vsTeam := e.Game.HomeTeamName
	if e.Player.TeamFk == e.Game.HomeTeamFk {
		vsTeam = e.Game.AwayTeamName
	}
	return fmt.Sprintf("%s.%s (%s-%s) %s vs %s",
		e.Player.FirstName[:1],
		e.Player.LastName,
		e.Player.TeamShort,
		e.Player.Position,
		e.Metric.Name,
		vsTeam,
	)
}

func (e PlayerExpression) ResultString() string {
	if e.Value == nil {
		return e.String()
	}
	return fmt.Sprintf("%s (%s)", e.String(), fmt.Sprintf("%.2f", *e.Value))
}
