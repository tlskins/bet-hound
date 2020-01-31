package types

import (
	"fmt"
	"strings"
)

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
