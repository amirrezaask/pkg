package httphandlers

import (
	"fmt"
	"log/slog"
	"net/http"

	json "github.com/json-iterator/go"
)

func DecodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func MakeHTTPHandler(f func(w http.ResponseWriter, r *http.Request) (int, any, error)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, resp, err := f(w, r)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			if status == 0 {
				status = http.StatusInternalServerError
			}

			w.WriteHeader(status)

			if resp == nil {
				w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, http.StatusText(status))))
			}
			slog.Error("error in http handler",
				"uri", r.URL.Path,
				"query", r.URL.RawQuery,
				"err", err,
			)

		} else {
			if status == 0 {
				status = 200
			}
			w.WriteHeader(status)
		}
		if resp != nil {
			err = json.NewEncoder(w).Encode(resp)
			if err != nil {
				slog.Error("error in writing response to ResponseWriter", "err", err)
			}
		}

	})
}
