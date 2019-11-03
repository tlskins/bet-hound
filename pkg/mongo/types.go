package mongo

import (
	"github.com/globalsign/mgo/bson"
	"github.com/shopspring/decimal"
)

type M map[string]interface{}

type Decimal decimal.Decimal

func NewDecimal(num int64, exp int32) Decimal {
	return Decimal(decimal.New(num, exp))
}

func NewDecimalFromString(s string) (d Decimal, err error) {
	var dec decimal.Decimal
	if dec, err = decimal.NewFromString(s); err != nil {
		return d, err
	}
	return Decimal(dec), err
}

func (d Decimal) GetBSON() (interface{}, error) {
	return bson.ParseDecimal128(decimal.Decimal(d).String())
}
func (d *Decimal) SetBSON(raw bson.Raw) (err error) {
	var dec128 bson.Decimal128
	if err = raw.Unmarshal(&dec128); err != nil {
		return err
	}
	var dec decimal.Decimal
	if dec, err = decimal.NewFromString(dec128.String()); err != nil {
		return err
	}
	*d = Decimal(dec)
	return nil
}
func (d Decimal) Abs() Decimal {
	return Decimal(decimal.Decimal(d).Abs())
}
func (d Decimal) Add(d2 Decimal) Decimal {
	return Decimal(decimal.Decimal(d).Add(decimal.Decimal(d2)))
}
func (d Decimal) Ceil() Decimal {
	return Decimal(decimal.Decimal(d).Ceil())
}
func (d Decimal) Div(d2 Decimal) Decimal {
	return Decimal(decimal.Decimal(d).Div(decimal.Decimal(d2)))
}
func (d Decimal) DivRound(d2 Decimal, precision int32) Decimal {
	return Decimal(decimal.Decimal(d).DivRound(decimal.Decimal(d2), precision))
}
func (d Decimal) Equal(d2 Decimal) bool {
	return decimal.Decimal(d).Equal(decimal.Decimal(d2))
}
func (d Decimal) Float64() (f float64, exact bool) {
	return decimal.Decimal(d).Float64()
}
func (d Decimal) Floor() Decimal {
	return Decimal(decimal.Decimal(d).Floor())
}
func (d Decimal) GreaterThan(d2 Decimal) bool {
	return decimal.Decimal(d).GreaterThan(decimal.Decimal(d2))
}
func (d Decimal) GreaterThanOrEqual(d2 Decimal) bool {
	return decimal.Decimal(d).GreaterThanOrEqual(decimal.Decimal(d2))
}
func (d Decimal) IsNegative() bool {
	return decimal.Decimal(d).IsNegative()
}
func (d Decimal) IsPositive() bool {
	return decimal.Decimal(d).IsPositive()
}
func (d Decimal) IsZero() bool {
	return decimal.Decimal(d).IsZero()
}
func (d Decimal) LessThan(d2 Decimal) bool {
	return decimal.Decimal(d).LessThan(decimal.Decimal(d2))
}
func (d Decimal) LessThanOrEqual(d2 Decimal) bool {
	return decimal.Decimal(d).LessThanOrEqual(decimal.Decimal(d2))
}
func (d Decimal) MarshalJSON() ([]byte, error) {
	return decimal.Decimal(d).MarshalJSON()
}
func (d Decimal) Mul(d2 Decimal) Decimal {
	return Decimal(decimal.Decimal(d).Mul(decimal.Decimal(d2)))
}
func (d Decimal) Round(places int32) Decimal {
	return Decimal(decimal.Decimal(d).Round(places))
}
func (d Decimal) RoundBank(places int32) Decimal {
	return Decimal(decimal.Decimal(d).RoundBank(places))
}
func (d Decimal) RoundCash(interval uint8) Decimal {
	return Decimal(decimal.Decimal(d).RoundCash(interval))
}
func (d Decimal) Sign() int {
	return decimal.Decimal(d).Sign()
}
func (d Decimal) String() string {
	return decimal.Decimal(d).String()
}
func (d Decimal) Truncate(precision int32) Decimal {
	return Decimal(decimal.Decimal(d).Truncate(precision))
}
func (d *Decimal) UnmarshalJSON(decimalBytes []byte) error {
	return (*decimal.Decimal)(d).UnmarshalJSON(decimalBytes)
}
