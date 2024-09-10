package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/getsentry/sentry-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// jsonIterSerializer implements JSON encoding using encoding/json.
type jsonIterSerializer struct{}

// Serialize converts an interface into a json and writes it to the response.
// You can optionally use the indent parameter to produce pretty JSONs.
func (d jsonIterSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := jsoniter.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

// Deserialize reads a JSON from a request body and converts it into an interface.
func (d jsonIterSerializer) Deserialize(c echo.Context, i interface{}) error {
	err := jsoniter.NewDecoder(c.Request().Body).Decode(i)
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).SetInternal(err)
	} else if se, ok := err.(*json.SyntaxError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
	}
	return err
}

func setTracingSpanMiddleware(service string) echo.MiddlewareFunc {
	return otelecho.Middleware(service)
}

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
