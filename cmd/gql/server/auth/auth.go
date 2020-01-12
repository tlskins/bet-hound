package auth

import (
	"context"
	"fmt"
	"net/http"
)

type ContextKey string

type AuthResponseWriter struct {
	http.ResponseWriter
	UserId string
}

func (w *AuthResponseWriter) SetSession(userId string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    userId,
		HttpOnly: true,
		Path:     "/",
		Domain:   "localhost",
	})
}

// func (w *AuthResponseWriter) Write(b []byte) (int, error) {
// 	fmt.Println("write=", w.userIDToResolver, w.userIDFromCookie)
// 	if w.userIDToResolver != w.userIDFromCookie {
// 		fmt.Println("setting cookie=", w.userIDToResolver, w.userIDFromCookie)
// 		http.SetCookie(w, &http.Cookie{
// 			Name:     "auth",
// 			Value:    w.userIDToResolver,
// 			HttpOnly: true,
// 			Path:     "/",
// 			Domain:   "localhost",
// 		})
// 	}
// 	// fmt.Println("setting cookie=", "test")
// 	// http.SetCookie(w, &http.Cookie{
// 	// 	Name:     "auth",
// 	// 	Value:    "Test",
// 	// 	HttpOnly: true,
// 	// 	Path:     "/",
// 	// 	Domain:   "localhost",
// 	// })
// 	return w.ResponseWriter.Write(b)
// }

func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		arw := AuthResponseWriter{w, ""}
		userIDContextKey := ContextKey("userID")

		c, err := r.Cookie("auth")
		fmt.Println("get auth cookie", c, err)
		if c != nil {
			fmt.Println("cookie value=", c.Value)
			arw.UserId = c.Value
		}
		// ctx := context.WithValue(r.Context(), userIDContextKey, &arw.userIDToResolver)
		ctx := context.WithValue(r.Context(), userIDContextKey, &arw)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
		arw.Write([]byte(""))
	})
}

// func AuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		arw := authResponseWriter{w, "", ""}
// 		userIDContextKey := gql.ContextKey("userID")

// 		c, err := r.Cookie("auth")
// 		fmt.Println("get auth cookie", c, err)
// 		if c != nil {
// 			fmt.Println("cookie=", *c)
// 			arw.userIDFromCookie = c.Value
// 			arw.userIDToResolver = c.Value
// 		}
// 		ctx := context.WithValue(r.Context(), userIDContextKey, &arw.userIDToResolver)
// 		r = r.WithContext(ctx)

// 		arw.Write([]byte(""))
// 		next.ServeHTTP(w, r)
// 	})
// }

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	"bet-hound/cmd/db"
// 	"bet-hound/cmd/gql"
// )

// type GqlAuth struct {
// 	UserName string `json:"userName"`
// 	Password string `json:"password"`
// }

// type GqlSignInReq struct {
// 	OperationName string  `json:"operationName"`
// 	Auth          GqlAuth `json:"variables"`
// 	Query         string  `json:"query"`
// }

// func AuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		auth := r.Header.Get("Authorization")
// 		if auth != "" {
// 			// Write your fancy token introspection logic here and if valid user then pass appropriate key in header
// 			// IMPORTANT: DO NOT HANDLE UNAUTHORISED USER HERE
// 			ctx = context.WithValue(ctx, gql.UserIDCtxKey, auth)
// 		} else {
// 			// sign in user
// 			var gqlRequest GqlSignInReq
// 			_ = json.NewDecoder(r.Body).Decode(&gqlRequest)
// 			if gqlRequest.OperationName == "signIn" {
// 				fmt.Println("gqlrequest", gqlRequest)
// 				user, err := db.SignInUser(gqlRequest.Auth.UserName, gqlRequest.Auth.Password)
// 				fmt.Println("user, err", user, err)
// 				if err == nil {
// 					fmt.Println("signed in")
// 					w.Header().Set("content-type", "application/json")
// 					w.Write([]byte(`{ "Authorization": "` + user.Id + `" }`))
// 				}
// 			}
// 		}
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }
