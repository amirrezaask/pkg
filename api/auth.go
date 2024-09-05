package api

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const (
	USER_CONTEXT_KEY = "user"
)

type AuthClaim struct {
	jwt.RegisteredClaims
	Subject int    `json:"sub,omitempty"`
	Mobile  string `json:"mobile"`
}

func AuthenticateMiddleware(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authorization := c.Request().Header.Get("Authorization")
			authorization, ok := strings.CutPrefix(authorization, "Bearer ")
			if !ok {
				return c.JSON(http.StatusUnauthorized, "unauthorized access")
			}

			tok, err := jwt.ParseWithClaims(authorization, &AuthClaim{}, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				return c.JSON(http.StatusUnauthorized, "unauthorized access")
			}

			c.Set(USER_CONTEXT_KEY, tok.Claims)
			return next(c)
		}
	}
}

func OptionalAuthenticate(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authorization := c.Request().Header.Get("Authorization")
			authorization, ok := strings.CutPrefix(authorization, "Bearer ")
			if !ok {
				return next(c)
			}

			tok, err := jwt.ParseWithClaims(authorization, &AuthClaim{}, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				return next(c)
			}

			c.Set(USER_CONTEXT_KEY, tok.Claims)
			return next(c)
		}

	}
}
