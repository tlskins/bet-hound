package middleware

import (
	"context"
	"net/http"

	t "bet-hound/cmd/types"
)

type LgContextKey string

func LeagueMiddleWare(next http.Handler, lgSttgs *t.LeagueSettings) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgContextKey := LgContextKey("league")

		ctx := context.WithValue(r.Context(), lgContextKey, lgSttgs)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
