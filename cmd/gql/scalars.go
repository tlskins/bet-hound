package gql

import (
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"

	"bet-hound/cmd/types"
)

func MarshalBetStatus(t types.BetStatus) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, "\""+t.String()+"\"")
	})
}

func UnmarshalBetStatus(v interface{}) (types.BetStatus, error) {
	if tmpStr, ok := v.(string); ok {
		return types.BetStatusFromString(tmpStr), nil
	}
	return types.BetStatusFromString("Pending Approval"), errors.New("bet status invalid")
}

func MarshalTimestamp(t time.Time) graphql.Marshaler {
	timestamp := t.Unix()
	if timestamp < 0 {
		timestamp = 0
	}
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatInt(timestamp, 10))
	})
}

func UnmarshalTimestamp(v interface{}) (time.Time, error) {
	if tmpStr, ok := v.(int); ok {
		return time.Unix(int64(tmpStr), 0), nil
	}
	return time.Time{}, errors.New("time should be a unix timestamp")
}
