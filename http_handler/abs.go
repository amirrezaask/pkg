package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response[T any] struct {
	Body   T
	Status int
	Err    error
}

var EchoBodyDecoder = func(req echo.Context, out any) error {
	return json.NewDecoder(req.Request().Body).Decode(out)
}

var StdBodyDecoder = func(req *http.Request, out any) error {
	return json.NewDecoder(req.Body).Decode(out)
}

func EchoHandler[IN, OUT any](handler func(c echo.Context, body IN) (Response[OUT], error)) echo.HandlerFunc {
	reqDecoder := EchoBodyDecoder
	respErrEncoder := json.Marshal
	respEncoder := json.Marshal
	return func(c echo.Context) error {
		var req IN
		err := reqDecoder(c, req)
		if err != nil {
			return err
		}
		resp, err := handler(c, req)
		c.Response().WriteHeader(resp.Status)
		if err != nil {
			bs, _ := respErrEncoder(err)
			c.Response().Write(bs)
			return nil
		}
		bs, _ := respEncoder(resp.Body)
		c.Response().Write(bs)
		return nil
	}
}

func StdHandler[IN, OUT any](handler func(r *http.Request, body IN) (Response[OUT], error)) http.HandlerFunc {
	reqDecoder := StdBodyDecoder
	respErrEncoder := json.Marshal
	respEncoder := json.Marshal
	return func(w http.ResponseWriter, r *http.Request) {
		var body IN
		err := reqDecoder(r, body)
		if err != nil {
			bs, _ := respErrEncoder(err)
			w.WriteHeader(400)
			w.Write(bs)
		}
		resp, err := handler(r, body)
		w.WriteHeader(resp.Status)
		if err != nil {
			bs, _ := respErrEncoder(err)
			w.Write(bs)
			return
		}
		bs, _ := respEncoder(resp.Body)
		w.Write(bs)
	}
}
