package httphandler

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
)

var EchoErrorEncoder func(err any) ([]byte, error) = json.Marshal
var EchoRespEncoder func(t any) ([]byte, error) = json.Marshal
var EchoReqDecoder func(c echo.Context, t any) error = func(c echo.Context, t any) error {
	return json.NewDecoder(c.Request().Body).Decode(t)
}

func MakeEchoHandler[IN, OUT any](handler func(Request[IN]) (Response[OUT], error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req IN
		err := json.NewDecoder(c.Request().Body).Decode(req)
		if err != nil {
			return err
		}
		resp, err := handler(Request[IN]{Body: req, Headers: c.Request().Header, Query: c.Request().URL.Query()})
		c.Response().WriteHeader(resp.Status)
		if err != nil {
			bs, _ := StdErrorEncoder(err)
			c.Response().Write(bs)
			return nil
		}
		bs, _ := StdRespEncoder(resp.Body)
		c.Response().Write(bs)
		return nil
	}
}

type User struct{}

func getuser(r Request[User]) (Response[User], error) {
	return Response[User]{}, nil
}

func main() {
	e := echo.New()
	e.GET("", MakeEchoHandler(getuser))
}
