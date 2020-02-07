package gql

import (
	"errors"
	"io"
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
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, "\""+t.String()+"\"")
	})
}

func UnmarshalTimestamp(v interface{}) (t time.Time, err error) {
	if tmpStr, ok := v.(string); ok {
		return time.Parse(
			time.RFC3339,
			tmpStr,
		)
	}
	return time.Time{}, errors.New("time should be in RFC3339 format")
}
