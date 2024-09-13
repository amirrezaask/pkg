package middlewares

import (
	"log/slog"
	"net/http"
)

func Recover(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				if err, isErr := r.(error); isErr {
					slog.Error(err.Error(), "panic", true)
				} else {
					slog.Error("unknown recover", "recover()", r)
				}
			}
		}()
	}
}
