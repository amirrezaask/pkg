package httphandler

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

type User struct{}
type UserResponse struct{}

func getUserHandlerEcho(c echo.Context, body User) (Response[User], error) {
	return Response[User]{}, nil
}

func TestEcho(t *testing.T) {
	e := echo.New()

	e.GET("/users", EchoHandler(getUserHandlerEcho))
}

func getUserHandlerStd(r *http.Request, body User) (Response[User], error) {
	return Response[User]{}, nil
}

func TestStd(t *testing.T) {
	http.HandleFunc("", StdHandler(getUserHandlerStd))
}
