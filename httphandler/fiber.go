package httphandler

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

var FiberCallbacks = Callbacks[*fiber.Ctx]{
	BodyDecoder: func(req *fiber.Ctx, out any) error {
		return json.Unmarshal(req.Request().Body(), out)
	},
}

func FiberHandler[IN any](handler func(c *fiber.Ctx, body IN) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req IN
		_ = FiberCallbacks.BodyDecoder(c, req)
		return handler(c, req)
	}
}
