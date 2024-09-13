package api

import (
	"log/slog"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

type Option func(e *echo.Echo)

func StartServer(listenAddr string,
	options ...func(e *echo.Echo),
) error {
	e := echo.New()
	e.JSONSerializer = &jsonIterSerializer{}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				r := recover()
				if r != nil {
					sentry.CurrentHub().Recover(r)
					if err, isErr := r.(error); isErr {
						slog.Error(err.Error(), "panic", true)
					} else {
						slog.Error("unknown recover", "recover()", r)
					}
				}
			}()

			return next(c)
		}
	})

	for _, opt := range options {
		opt(e)
	}

	return e.Start(listenAddr)
}

type Request[T any] struct {
	echo.Context
	Binding T
}

type Response struct {
	StatusCode int
	Body       any
}

func Wrap[IN any](f func(Request[IN]) Response) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in IN
		err := c.Bind(&in)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"message": "bad request"})
		}

		resp := f(Request[IN]{Context: c, Binding: in})
		if resp.StatusCode == 0 {
			resp.StatusCode = 200
		}

		return c.JSON(resp.StatusCode, resp.Body)
	}
}
