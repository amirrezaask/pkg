package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4"
)

type Callbacks[Req any] struct {
	BodyDecoder     func(req Req, out any) error
	RespBodyEncoder func(v any) ([]byte, error)
	RespErrEncoder  func(v any) ([]byte, error)
}

var EchoCallbacks = Callbacks[echo.Context]{
	BodyDecoder: func(req echo.Context, out any) error {
		return json.NewDecoder(req.Request().Body).Decode(out)
	},
	RespBodyEncoder: json.Marshal,
	RespErrEncoder:  json.Marshal,
}

var StdCallbacks = Callbacks[*http.Request]{
	BodyDecoder: func(req *http.Request, out any) error {
		return json.NewDecoder(req.Body).Decode(out)
	},
	RespBodyEncoder: json.Marshal,
	RespErrEncoder:  json.Marshal,
}

var GinCallbacks = Callbacks[*gin.Context]{
	BodyDecoder: func(req *gin.Context, out any) error {
		return json.NewDecoder(req.Request.Body).Decode(out)
	},
	RespBodyEncoder: json.Marshal,
	RespErrEncoder:  json.Marshal,
}

var FiberCallbacks = Callbacks[*fiber.Ctx]{
	BodyDecoder: func(req *fiber.Ctx, out any) error {
		return json.Unmarshal(req.Request().Body(), out)
	},
	RespBodyEncoder: json.Marshal,
	RespErrEncoder:  json.Marshal,
}

func EchoHandler[IN, OUT any](handler func(c echo.Context, body IN) (int, http.Header, OUT, error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req IN
		err := EchoCallbacks.BodyDecoder(c, req)
		if err != nil {
			return err
		}
		status, headers, body, err := handler(c, req)
		c.Response().WriteHeader(status)
		for k, vs := range headers {
			for _, v := range vs {
				c.Response().Header().Add(k, v)
			}
		}
		if err != nil {
			bs, _ := EchoCallbacks.RespErrEncoder(err)
			c.Response().Write(bs)
			return nil
		}
		bs, _ := EchoCallbacks.RespBodyEncoder(body)
		c.Response().Write(bs)
		return nil
	}
}

func StdHandler[IN, OUT any](handler func(r *http.Request, body IN) (int, http.Header, OUT, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body IN
		err := StdCallbacks.BodyDecoder(r, body)
		if err != nil {
			bs, _ := StdCallbacks.RespErrEncoder(err)
			w.WriteHeader(400)
			w.Write(bs)
		}
		status, headers, respBody, err := handler(r, body)
		w.WriteHeader(status)
		for k, vs := range headers {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		if err != nil {
			bs, _ := StdCallbacks.RespErrEncoder(err)
			w.Write(bs)
			return
		}
		bs, _ := StdCallbacks.RespBodyEncoder(respBody)
		w.Write(bs)
	}
}

func GinHandler[IN, OUT any](handler func(c *gin.Context, body IN) (int, http.Header, OUT, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req IN
		_ = GinCallbacks.BodyDecoder(c, req)
		status, headers, body, err := handler(c, req)
		c.Writer.WriteHeader(status)
		for k, vs := range headers {
			for _, v := range vs {
				c.Writer.Header().Add(k, v)
			}
		}
		if err != nil {
			bs, _ := EchoCallbacks.RespErrEncoder(err)
			c.Writer.Write(bs)
			return
		}
		bs, _ := EchoCallbacks.RespBodyEncoder(body)
		c.Writer.Write(bs)
	}
}

func FiberHandler[IN, OUT any](handler func(c *fiber.Ctx, body IN) (int, http.Header, OUT, error)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req IN
		_ = FiberCallbacks.BodyDecoder(c, req)
		status, headers, body, err := handler(c, req)
		c.Response().SetStatusCode(status)
		for k, vs := range headers {
			for _, v := range vs {
				c.Response().Header.Add(k, v)
			}
		}
		if err != nil {
			bs, _ := EchoCallbacks.RespErrEncoder(err)
			c.Response().BodyWriter().Write(bs)
			return nil
		}
		bs, _ := EchoCallbacks.RespBodyEncoder(body)
		c.Response().BodyWriter().Write(bs)
		return nil
	}
}
