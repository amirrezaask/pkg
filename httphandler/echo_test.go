package httphandler

import (
	"testing"

	"github.com/labstack/echo/v4"
)

type User struct {
	Name string `json:"name"`
}

func getUserHandlerEcho(c echo.Context, body User) error {
	return nil
}

func TestEcho(t *testing.T) {
	e := echo.New()

	e.GET("/users", EchoHandler(getUserHandlerEcho))
}
