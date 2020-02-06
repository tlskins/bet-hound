package types

// Changes

type NewBet struct {
	LeagueId     string         `json:"league_id"`
	BetRecipient BetRecipient   `json:"recipient"`
	NewEquations []*NewEquation `json:"new_equations"`
}

type BetRecipient struct {
	UserId            *string `json:"user_id"`
	TwitterScreenName *string `json:"twitter_screen_name"`
}

type NewEquation struct {
	OperatorId     *int             `json:"operator_id"`
	NewExpressions []*NewExpression `json:"new_expressions"`
}

type NewExpression struct {
	IsLeft   bool     `json:"is_left"`
	PlayerId *string  `json:"player_id"`
	GameId   *string  `json:"game_id"`
	TeamId   *string  `json:"team_id"`
	MetricId *int     `json:"metric_id"`
	Value    *float64 `json:"value"`
}

// TODO : Add validations
