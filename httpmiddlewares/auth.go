package httpmiddlewares

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ClaimsKey          = "claims"
	IsAuthenticatedKey = "isAuthenticated"
)

func BasicAuth(user string, pass string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				h.ServeHTTP(w, r)
				return
			}

			if username == user && password == pass {
				r = r.WithContext(context.WithValue(r.Context(), IsAuthenticatedKey, true))
				r = r.WithContext(context.WithValue(r.Context(), ClaimsKey, username))
			}

			h.ServeHTTP(w, r)
		})
	}
}

func BearerAuth[C any, CLAIMS interface {
	jwt.Claims
	*C
}](secret []byte) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			authorization, ok := strings.CutPrefix(authorization, "Bearer ")
			if !ok {
				h.ServeHTTP(w, r)
				return
			}
			claims := CLAIMS(new(C))
			tok, err := jwt.ParseWithClaims(authorization, claims, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				slog.Error("error in parsing jwt with claims", "err", err)
				h.ServeHTTP(w, r)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), ClaimsKey, tok.Claims))
			r = r.WithContext(context.WithValue(r.Context(), IsAuthenticatedKey, true))
			h.ServeHTTP(w, r)
		})
	}
}

func AuthenticatedOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(IsAuthenticatedKey) != true {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized user"))
			return
		}
		h.ServeHTTP(w, r)
	})
}
