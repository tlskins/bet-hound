package auth

import (
	"context"
	"net/http"
)

type ContextKey string

type AuthResponseWriter struct {
	http.ResponseWriter
	UserId string
}

func (w *AuthResponseWriter) SetSession(appUrl, userId string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    userId,
		HttpOnly: true,
		Path:     "/",
		Domain:   appUrl,
	})
}

func (w *AuthResponseWriter) DeleteSession(appUrl string) bool {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   0,
		Domain:   appUrl,
	})
	return true
}

func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

		arw := AuthResponseWriter{w, ""}
		userIDContextKey := ContextKey("userID")

		c, _ := r.Cookie("auth")
		if c != nil {
			arw.UserId = c.Value
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, &arw)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
		arw.Write([]byte(""))
	})
}
