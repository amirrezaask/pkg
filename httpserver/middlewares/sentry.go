package middlewares

import (
	"net/http"

	"github.com/getsentry/sentry-go"
)

func Sentry(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				sentry.CurrentHub().Recover(r)
				panic(r)
			}
		}()
	}
}
