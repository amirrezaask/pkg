package httphandler

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/labstack/echo/v4"
)

type User struct {
	Name string `json:"name"`
}

func getUserEcho(c echo.Context) error {
	var user User
	err := c.Bind(&user)
	if err != nil {
		return c.NoContent(400)
	}

	// logic
	fmt.Println(user.Name)
	return nil
}

func getUserHandlerEcho(c echo.Context, body User) error {
	fmt.Println(body.Name)
	return nil
}

func TestEcho(t *testing.T) {
	e := echo.New()

	EchoCallbacks.BodyDecoder = func(req echo.Context, out any) error {
		return xml.NewDecoder(req.Request().Body).Decode(out)
	}

	e.GET("/usrs", getUserEcho)
	e.GET("/users", EchoHandler(getUserHandlerEcho))
}
