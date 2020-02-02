package types

import (
	"fmt"
	"strings"
)

// for unmarshalling then converting to bet

type MongoEquation struct {
	Id          int               `bson:"id" json:"id"`
	Expressions []MongoExpression `bson:"exprs" json:"expressions"`
	Operator    *BetMap           `bson:"op" json:"operator"`
	Result      *bool             `bson:"res" json:"result"`
}

func (m MongoEquation) Equation() *Equation {
	exps := make([]Expression, len(m.Expressions))
	for i, exp := range m.Expressions {
		exps[i] = exp.Expression()
	}

	return &Equation{
		Id:          m.Id,
		Operator:    m.Operator,
		Result:      m.Result,
		Expressions: exps,
	}
}

// Equation

type Equation struct {
	Id          int          `bson:"id" json:"id"`
	Expressions []Expression `bson:"exprs" json:"expressions"`
	Operator    *BetMap      `bson:"op" json:"operator"`
	Result      *bool        `bson:"res" json:"result"`
}

func (e Equation) LeftExpressions() (exprs []Expression) {
	exprs = []Expression{}
	for _, expr := range e.Expressions {
		if expr.IsLeft() {
			exprs = append(exprs, expr)
		}
	}
	return
}

func (e Equation) RightExpressions() (exprs []Expression) {
	exprs = []Expression{}
	for _, expr := range e.Expressions {
		if !expr.IsLeft() {
			exprs = append(exprs, expr)
		}
	}
	return
}

func (e Equation) Valid() error {
	left := e.LeftExpressions()
	if len(left) == 0 {
		return fmt.Errorf("No left expressions found.")
	} else if left[0].GetGame() == nil {
		return fmt.Errorf("First left expression must have a game.")
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
		if expr.IsLeft() {
			left = append(left, expr.ResultString())
		} else if len(expr.ResultString()) > 0 {
			right = append(right, expr.ResultString())
		}
	}

	if len(right) > 0 {
		return fmt.Sprintf(
			"%s %s %s",
			strings.Join(left, " "),
			e.Operator.Name,
			strings.Join(right, " "),
		)
	} else {
		return strings.Join(left, " ")
	}
}

func (e Equation) ResultString() string {
	if e.Result == nil {
		return e.String()
	}
	return fmt.Sprintf("%s (%t)", e.String(), *e.Result)
}
