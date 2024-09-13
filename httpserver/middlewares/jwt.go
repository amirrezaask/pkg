package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	JWTClaimsKey = "user"
)

func ParseJWT[C any, CLAIMS interface {
	jwt.Claims
	*C
}](secret []byte) func(h http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			authorization, ok := strings.CutPrefix(authorization, "Bearer ")
			if !ok {
				h(w, r)
			}
			claims := CLAIMS(new(C))
			tok, err := jwt.ParseWithClaims(authorization, claims, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				h(w, r)
			}
			r = r.WithContext(context.WithValue(r.Context(), JWTClaimsKey, tok.Claims))
			h(w, r)
		}
	}

}

func AuthenticatedOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(JWTClaimsKey) == nil {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized user"))
			return
		}

		h(w, r)
	}
}
