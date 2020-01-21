package middleware

import (
	"context"
	"net/http"
	"strings"
)

type AuthContextKey string

type AuthResponseWriter struct {
	http.ResponseWriter
	UserId string
}

func (w *AuthResponseWriter) SetSession(appHost, userId string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    userId,
		HttpOnly: true,
		Path:     "/",
		Domain:   appHost,
	})
}

func (w *AuthResponseWriter) DeleteSession(appHost string) bool {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   0,
		Domain:   appHost,
	})
	return true
}

func AuthMiddleWare(next http.Handler, allowOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		for _, origin := range strings.Split(allowOrigin, ",") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		arw := AuthResponseWriter{w, ""}
		userIDAuthContextKey := AuthContextKey("userID")

		c, _ := r.Cookie("auth")
		if c != nil {
			arw.UserId = c.Value
		}
		ctx := context.WithValue(r.Context(), userIDAuthContextKey, &arw)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
		arw.Write([]byte(""))
	})
}
