package types

// Bet Maps

type BetMap struct {
	Id                   int       `bson:"id" json:"id"`
	LeagueId             string    `bson:"lg_id" json:"league_id"`
	Type                 string    `bson:"t" json:"type"`
	Name                 string    `bson:"n" json:"name"`
	Field                string    `bson:"f" json:"field"`
	LeftOnly             bool      `bson:"lft" json:"left_only"`
	OperatorId           *int      `bson:"op_id" json:"operator_id"`
	RightExpressionTypes *[]string `bson:"rgt_tps" json:"right_expression_types"`
	RightExpressionValue *float64  `bson:"rgt_vl" json:"right_expression_value"`
}
