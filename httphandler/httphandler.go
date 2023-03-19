package httphandler

import (
	"encoding/json"
	"net/http"
)

type Callbacks[Req any] struct {
	BodyDecoder func(req Req, out any) error
}

type Response struct {
	Status  int
	Headers http.Header
	Body    any
}

var StdCallbacks = Callbacks[*http.Request]{
	BodyDecoder: func(req *http.Request, out any) error {
		return json.NewDecoder(req.Body).Decode(out)
	},
}

func StdHandler[IN any](handler func(r *http.Request, body IN)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body IN
		err := StdCallbacks.BodyDecoder(r, body)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		handler(r, body)
	}
}
