package httphandler

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
)

var EchoCallbacks = Callbacks[echo.Context]{
	BodyDecoder: func(req echo.Context, out any) error {
		return json.NewDecoder(req.Request().Body).Decode(out)
	},
}

func EchoHandler[IN any](handler func(c echo.Context, body IN) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req IN
		err := EchoCallbacks.BodyDecoder(c, req)
		if err != nil {
			return err
		}
		return handler(c, req)
	}
}
