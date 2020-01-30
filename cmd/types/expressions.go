package types

import "fmt"

// Static Expressions

type StaticExpression struct {
	Id    int      `bson:"id" json:"id"`
	Value *float64 `bson:"val" json:"value"`
}

func (e StaticExpression) ResultValue() *float64 {
	return e.Value
}

func (e StaticExpression) IsLeft() bool {
	return false
}

func (e StaticExpression) Valid() error {
	if e.Value == nil {
		return fmt.Errorf("Value not found for static expression")
	} else {
		return nil
	}
}

func (e StaticExpression) String() (desc string) {
	if e.Value == nil {
		return "?"
	}
	return fmt.Sprintf("%.0f", *e.Value)
}

func (e StaticExpression) ResultString() string {
	return e.String()
}

// Team Expressions

type TeamExpression struct {
	Id     int      `bson:"id" json:"id"`
	Left   bool     `bson:"lft" json:"is_left"`
	Team   *Team    `bson:"gm" json:"game"`
	Value  *float64 `bson:"val" json:"value"`
	Metric *BetMap  `bson:"mtc" json:"metric"`
}

func (e TeamExpression) ResultValue() *float64 {
	return e.Value
}

func (e TeamExpression) IsLeft() bool {
	return e.Left
}

func (e TeamExpression) Valid() error {
	if e.Team == nil {
		return fmt.Errorf("Team not found")
	} else if e.Metric == nil {
		return fmt.Errorf("Metric not found for team")
	} else {
		return nil
	}
}

func (e TeamExpression) String() (desc string) {
	if e.Team == nil || e.Metric == nil {
		return "?"
	}
	return fmt.Sprintf("%s (%s)", e.Team.Name, e.Team.Location)
}

func (e TeamExpression) ResultString() string {
	if e.Value == nil {
		return e.String()
	}
	return fmt.Sprintf("%s (%s)", e.String(), fmt.Sprintf("%.2f", *e.Value))
}

// Player Expressions

type PlayerExpression struct {
	Id     int      `bson:"id" json:"id"`
	IsLeft bool     `bson:"lft" json:"is_left"`
	Team   *Team    `bson:"tm" json:"team,omitempty"`
	Player *Player  `bson:"player" json:"player,omitempty"`
	Game   *Game    `bson:"gm" json:"game,omitempty"`
	Value  *float64 `bson:"val" json:"value,omitempty"`
	Metric *BetMap  `bson:"mtc" json:"metric,omitempty"`
}

// func (e Expression) Type() (string, error) {
// 	if e.Team == nil && e.Player == nil && e.Game == nil && e.Value != nil {
// 		return "Static", nil
// 	}
// 	if e.Team != nil && e.Player == nil && e.Game != nil {
// 		return "Team", nil
// 	}
// 	if e.Player != nil && e.Game != nil {
// 		return "Player", nil
// 	}
// 	return "", fmt.Errorf("Incomplete expression")
// }

func (e PlayerExpression) ResultValue() *float64 {
	return e.Value
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
	if e.Player != nil && e.Game != nil {
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
