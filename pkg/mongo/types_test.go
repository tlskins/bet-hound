package mongo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecimal(t *testing.T) {
	type A struct {
		Id ObjectId `bson:"_id,omitempty"`
		A  Decimal
	}
	sess, _ := NewClient("127.0.0.1", "", "")
	defer sess.Close()
	c := sess.DB("hungry_test").C("pkg_mongo_test_decimal")
	d, _ := NewDecimalFromString("1.2")
	a := A{A: d}
	c.Insert(a)
	var as []A
	c.Find(nil).All(&as)
	fmt.Println(as[0])
	assert.True(t, as[0].A.Equal(d))
	c.DropCollection()
}
