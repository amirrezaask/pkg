package httpmiddlewares

import (
	"net/http"

	"github.com/getsentry/sentry-go"
)

func Sentry(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				sentry.CurrentHub().Recover(r)
				panic(r)
			}
		}()

		h.ServeHTTP(w, r)

	})
}
