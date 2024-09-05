package httpclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var prometheusDurationBuckets = []float64{
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

type transport struct {
	stdTransport         http.RoundTripper
	httpRequestDurationH *prometheus.HistogramVec
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()
	resp, err := t.stdTransport.RoundTrip(req)
	t.httpRequestDurationH.WithLabelValues(req.Method, req.URL.Host, req.URL.Path, resp.Status).Observe(time.Since(startTime).Seconds())

	return resp, err
}

func New(promNS string, name string, timeout time.Duration) *http.Client {
	t := &transport{
		stdTransport: http.DefaultTransport,
		httpRequestDurationH: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: promNS,
			Name:      fmt.Sprintf("http_client_%s_request_duration_seconds", name),
			Help:      fmt.Sprintf("Spend time for requests from %s client", name),
			Buckets:   prometheusDurationBuckets,
		},
			[]string{"method", "host", "uri", "status_code"}),
	}
	return &http.Client{Transport: t, Timeout: timeout}
}
