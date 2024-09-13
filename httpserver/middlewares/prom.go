package middlewares

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func PrometheusExporter(registry prometheus.Registerer, namespace string, excludePaths ...string) func(h http.HandlerFunc) http.HandlerFunc {
	pathRegexps := []*regexp.Regexp{}

	for _, path := range excludePaths {
		pathRegexps = append(pathRegexps, regexp.MustCompile(path))
	}
	Buckets := []float64{
		0.0005,
		0.001, // 1ms
		0.002,
		0.005,
		0.01, // 10ms
		0.02,
		0.05,
		0.1, // 100 ms
		0.2,
		0.5,
		1.0, // 1s
		2.0,
		5.0,
		10.0, // 10s
		15.0,
		20.0,
		30.0,
	}
	requestsHist := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "httpserver",
		Name:      "http_requests",
		Buckets:   Buckets,
	}, []string{"status", "method", "handler"})

	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			for _, path := range pathRegexps {
				if path.Match([]byte(req.RequestURI)) {
					h(w, req)
					return
				}
			}
			start := time.Now()
			statusRecorder := &statusRecorder{ResponseWriter: w}
			h(statusRecorder, req)
			requestsHist.WithLabelValues(fmt.Sprint(statusRecorder.status), req.Method, req.RequestURI).Observe(time.Since(start).Seconds())
		}
	}

}
