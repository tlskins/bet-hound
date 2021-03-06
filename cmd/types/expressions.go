package types

import (
	"fmt"
)

// for unmarshalling then converting to expression

type MongoExpression struct {
	Id     int      `bson:"id" json:"id"`
	Left   bool     `bson:"lft" json:"is_left"`
	Player *Player  `bson:"player" json:"player,omitempty"`
	Team   *Team    `bson:"tm" json:"team"`
	Game   *Game    `bson:"gm" json:"game,omitempty"`
	Value  *float64 `bson:"val" json:"value,omitempty"`
	Metric *BetMap  `bson:"mtc" json:"metric,omitempty"`
}

func (m MongoExpression) Expression() Expression {
	var exp Expression
	if m.Player != nil {
		exp = PlayerExpression{
			Id:     m.Id,
			Left:   m.Left,
			Player: m.Player,
			Game:   m.Game,
			Value:  m.Value,
			Metric: m.Metric,
		}
	} else if m.Team != nil {
		exp = TeamExpression{
			Id:     m.Id,
			Left:   m.Left,
			Team:   m.Team,
			Game:   m.Game,
			Value:  m.Value,
			Metric: m.Metric,
		}
	} else if m.Value != nil {
		exp = StaticExpression{
			Id:    m.Id,
			Value: m.Value,
		}
	}

	return exp
}

type Expression interface {
	ResultValue() *float64
	IsLeft() bool
	Valid() error
	String() string
	ResultString() string
	GetGame() *Game
}

type ExpressionUnion interface {
	ResultValue() *float64
	IsLeft() bool
	Valid() error
	String() string
	ResultString() string
	GetGame() *Game
}

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
	return ""
}

func (e StaticExpression) GetGame() *Game {
	return nil
}

// Team Expressions

type TeamExpression struct {
	Id     int      `bson:"id" json:"id"`
	Left   bool     `bson:"lft" json:"is_left"`
	Team   *Team    `bson:"tm" json:"team"`
	Game   *Game    `bson:"gm" json:"game,omitempty"`
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
	} else if e.Game == nil {
		return fmt.Errorf("Game not found for team")
	} else {
		return nil
	}
}

func (e TeamExpression) String() (desc string) {
	if e.Team == nil || e.Metric == nil {
		return "?"
	}
	return fmt.Sprintf("%s %s", e.Team.Name, e.Metric.Name)
}

func (e TeamExpression) ResultString() string {
	if e.Value == nil {
		return e.String()
	}
	return fmt.Sprintf("%s (%s)", e.String(), fmt.Sprintf("%.2f", *e.Value))
}

func (e TeamExpression) GetGame() *Game {
	return e.Game
}

// Player Expressions

type PlayerExpression struct {
	Id     int      `bson:"id" json:"id"`
	Left   bool     `bson:"lft" json:"is_left"`
	Player *Player  `bson:"player" json:"player,omitempty"`
	Game   *Game    `bson:"gm" json:"game,omitempty"`
	Value  *float64 `bson:"val" json:"value,omitempty"`
	Metric *BetMap  `bson:"mtc" json:"metric,omitempty"`
}

func (e PlayerExpression) IsLeft() bool {
	return e.Left
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
	if e.Player == nil && e.Game == nil {
		return "?"
	}
	return fmt.Sprintf("%s.%s %s",
		e.Player.FirstName[:1],
		e.Player.LastName,
		e.Metric.Name,
	)
}

func (e PlayerExpression) ResultString() string {
	if e.Value == nil {
		return e.String()
	}
	return fmt.Sprintf("%s (%s)", e.String(), fmt.Sprintf("%.2f", *e.Value))
}

func (e PlayerExpression) GetGame() *Game {
	return e.Game
}

func (e PlayerExpression) ResultValue() *float64 {
	return e.Value
}
