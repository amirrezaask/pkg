package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amirrezaask/pkg/errors"
)

type Request[T any] struct {
	*http.Request
	Binding T
}

type Response struct {
	StatusCode int
	Body       any
}

type RequestBinder func(r *http.Request, binding any) error
type Writer func(w http.ResponseWriter, status int, body any) error

func JSONBinder(r *http.Request, binding any) error {
	err := json.NewDecoder(r.Body).Decode(binding)
	if err != nil {
		return errors.Wrap(err, "error in json binder")
	}
	return nil
}

func JSONWriter(w http.ResponseWriter, status int, body any) error {
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(body)
}

func BinderHandler[T any](handler func(Request[T]) Response, binders ...any) http.HandlerFunc {
	var requestBinder RequestBinder
	var writer Writer
	switch {
	case len(binders) == 0:
		requestBinder = JSONBinder
		writer = JSONWriter
	case len(binders) == 1:
		requestBinder = binders[0].(RequestBinder)
		writer = JSONWriter
	default:
		requestBinder = binders[0].(RequestBinder)
		writer = binders[1].(Writer)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var binding T
		err := requestBinder(r, &binding)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"message": "bad request"}`)
			return
		}
		resp := handler(Request[T]{Request: r, Binding: binding})
		writer(w, resp.StatusCode, resp.Body)
	}
}
