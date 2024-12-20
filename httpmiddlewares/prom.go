package httpmiddlewares

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func PrometheusExporter(namespace string, excludePaths ...string) func(h http.Handler) http.Handler {
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
	requestsHist := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "httpserver",
		Name:      "requests_duration",
		Buckets:   Buckets,
	}, []string{"status", "method", "handler"})

	requestCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "httpserver",
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"status", "method", "handler"},
	)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			for _, path := range pathRegexps {
				if path.Match([]byte(req.RequestURI)) {
					h.ServeHTTP(w, req)
					return
				}
			}
			start := time.Now()
			statusRecorder := &statusRecorder{ResponseWriter: w}
			h.ServeHTTP(statusRecorder, req)

			requestCount.WithLabelValues(fmt.Sprint(statusRecorder.status), req.Method, req.URL.Path).
				Inc()

			requestsHist.WithLabelValues(fmt.Sprint(statusRecorder.status), req.Method, req.URL.Path).
				Observe(time.Since(start).Seconds())
		})
	}

}
