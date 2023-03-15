package httphandler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

type User struct{}
type UserResponse struct{}

func getUserHandlerEcho(c echo.Context, body User) (int, http.Header, UserResponse, error) {
	return 200, nil, UserResponse{}, nil
}

func TestEcho(t *testing.T) {
	e := echo.New()

	e.GET("/users", EchoHandler(getUserHandlerEcho))
}

func getUserHandlerStd(r *http.Request, body User) (int, http.Header, UserResponse, error) {
	return 200, nil, UserResponse{}, nil
}

func TestStd(t *testing.T) {
	http.HandleFunc("", StdHandler(getUserHandlerStd))
}

func getUserHandlerGin(c *gin.Context, body User) (int, http.Header, UserResponse, error) {
	return 200, nil, UserResponse{}, nil
}

func TestGin(t *testing.T) {
	g := gin.New()
	g.GET("", GinHandler(getUserHandlerGin))
}
