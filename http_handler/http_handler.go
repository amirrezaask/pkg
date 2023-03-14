package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func StdHTTP[T any](f func(r *http.Request) (T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			bs, _ := json.Marshal(err)
			fmt.Fprintf(w, "%s", string(bs))
			return
		}
		bs, _ := json.Marshal(resp)
		fmt.Fprintf(w, "%s", string(bs))
	}
}

func Echo[T any](f func(c echo.Context) (T, error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, err := f(c)
		if err != nil {
			c.Error(err)
			return nil
		}
		return c.JSON(200, resp)
	}
}
