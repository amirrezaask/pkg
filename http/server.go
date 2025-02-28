package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Result struct {
	Body   any
	Status int
	Header http.Header
}

type Request struct {
	*http.Request
}

func (r *Request) GetClaims() jwt.Claims {
	claims := r.Context().Value(ClaimsKey)
	if claims == nil {
		return nil
	}
	return claims.(jwt.Claims)
}

func (r *Request) Bind(v any) error {
	rvPtr := reflect.ValueOf(v)
	if rvPtr.Kind() != reflect.Ptr {
		return fmt.Errorf("input should be a pointer for Bind")
	}
	rv := rvPtr.Elem()
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		qp := rt.Field(i).Tag.Get("query")
		if qp == "" {
			continue
		}
		q := r.URL.Query().Get(qp)
		if q == "" {
			continue
		}
		setWithProperType(rt.Field(i).Type.Kind(), q, rv.Field(i))
	}
	for i := 0; i < rv.NumField(); i++ {
		pp := rt.Field(i).Tag.Get("path")
		if pp == "" {
			continue
		}
		q := r.PathValue(pp)
		if q == "" {
			continue
		}
		setWithProperType(rt.Field(i).Type.Kind(), q, rv.Field(i))
	}

	return r.BindBody(v)
}

func (r *Request) BindBody(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("input should be a pointer for Bind")
	}
	if r.Header.Get("Content-Type") == "" {
		r.Header.Set("Content-Type", "application/json")
	}
	switch r.Header.Get("Content-Type") {
	case "application/json":
		err := json.NewDecoder(r.Body).Decode(v)
		if err == io.EOF && r.Method == "GET" {
			return nil
		}

		return err
	default:
		return fmt.Errorf("Content-Type '%s' is not supported", r.Header.Get("Content-Type"))
	}
}

type HandlerFunc func(*Request) (Result, error)

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h(&Request{r})
	if err != nil {
		slog.Error("error in http handler", "err", err.Error())
	}

	if res.Status == 0 && err == nil {
		res.Status = 200
	} else if res.Status == 0 && err != nil {
		res.Status = 500
	}

	w.WriteHeader(res.Status)

	switch res.Body.(type) {
	case io.Reader:
		io.Copy(w, res.Body.(io.Reader))
	default:
		//marshal json
		json.NewEncoder(w).Encode(res.Body)
	}
}

type MiddlewareFunc = func(http.Handler) http.Handler
type ServeMux struct {
	*http.ServeMux
	middlewares []MiddlewareFunc
}

func NewServeMux() *ServeMux {
	return &ServeMux{ServeMux: http.NewServeMux()}
}

func (s *ServeMux) UseMiddlewares(middlewares ...MiddlewareFunc) {
	s.middlewares = append(s.middlewares, middlewares...)
}

func (s *ServeMux) MapPrometheusEndpoint(path string) {
	s.ServeMux.Handle("GET "+path, promhttp.Handler())
}

func (s *ServeMux) Handle(path string, handler http.Handler, middlewares ...MiddlewareFunc) {
	handler = ChainMiddlewares(append(s.middlewares, middlewares...)...)(handler)
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*r = *r.WithContext(context.WithValue(r.Context(), "registered_uri", path))
		handler.ServeHTTP(w, r)
	})
	s.ServeMux.Handle(path, wrapped)
}

func (s *ServeMux) handleFuncSimple(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	s.Handle(path, handler, middlewares...)
}

func (s *ServeMux) HandleFunc(path string, handler interface{}, middlewares ...MiddlewareFunc) {
	t := reflect.TypeOf(handler)
	v := reflect.ValueOf(handler)
	if t.Kind() != reflect.Func {
		panic("handler input of HandleFunc2 should be a function type")
	}

	switch handler := handler.(type) {
	case func(*Request) (Result, error):
		s.handleFuncSimple(path, handler, middlewares...)
		return
	case func(http.ResponseWriter, *Request):
		s.ServeMux.HandleFunc(path, func(rw http.ResponseWriter, r *http.Request) {
			defer _recover()
			handler(rw, &Request{r})
		})
		return
	}

	if t.NumIn() != 2 {
		panic(fmt.Sprintf("%T is not supported, input of HandleFunc should be either:\n%s\n%s\n%s\n", handler, "func(*http.Request) (Result, error)", "func(http.ResponseWriter, *http.Request)", "func(*http.Request, INPUTTYPE) (OUTPUTTYPE, error)"))
	}
	if t.In(0).String() != "*http.Request" {
		panic("first input of handler should be *http.Request, " + t.In(0).Name())
	}

	if t.Out(1).String() != "error" {
		panic("second output of handler should be error")
	}

	s.Handle(path, HandlerFunc(func(r *Request) (Result, error) {
		defer _recover()
		req := reflect.New(t.In(1).Elem())
		err := r.Bind(req.Interface())
		if err != nil {
			return Result{}, err
		}
		res := v.Call([]reflect.Value{reflect.ValueOf(r), req})
		errI := res[1].Interface()
		if errI == nil {
			return Result{Body: res[0].Interface()}, nil
		}
		return Result{Body: res[0].Interface()}, errI.(error)
	}), middlewares...)
}

func (s *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.ServeMux.ServeHTTP(w, r)
}

var DefaultServeMux = &ServeMux{}

func Handle(path string, handler http.Handler, middlewares ...MiddlewareFunc) {
	DefaultServeMux.Handle(path, handler, middlewares...)
}
func HandleFunc(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	DefaultServeMux.Handle(path, handler, middlewares...)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func PrometheusExporterMiddleware(namespace string, excludePaths ...string) func(h http.Handler) http.Handler {
	var pathRegexps []*regexp.Regexp
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
			rPath := req.Context().Value("registered_uri").(string)
			requestCount.WithLabelValues(fmt.Sprint(statusRecorder.status), req.Method, rPath).
				Inc()

			requestsHist.WithLabelValues(fmt.Sprint(statusRecorder.status), req.Method, rPath).
				Observe(time.Since(start).Seconds())
		})
	}

}

func _recover() {
	r := recover()
	if r != nil {
		stack := make([]byte, 4<<10)
		length := runtime.Stack(stack, true)
		stack = stack[:length]
		fmt.Println("PANIC STACK")
		fmt.Println(string(stack))
		if err, isErr := r.(error); isErr {
			slog.Error(err.Error(), "panic", true)
		} else {
			slog.Error("unknown recover",
				"recover()", r)
		}
	}
}

func RecoverMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer _recover()
		h.ServeHTTP(w, r)
	})
}

func RequestLoggerMiddleware(logWriter io.Writer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			statusRecorder := &statusRecorder{ResponseWriter: w}
			start := time.Now()
			next.ServeHTTP(statusRecorder, r)
			elapsed := time.Since(start)
			fmt.Fprintf(logWriter, "%s %s %s %d %s\n", r.Method, r.URL.Path, r.Proto, statusRecorder.status, elapsed)
		})
	}
}

func SentryMiddleware(h http.Handler) http.Handler {
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

var (
	ClaimsKey          = "claims"
	IsAuthenticatedKey = "isAuthenticated"
)

func BasicAuthMiddleware(user string, pass string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				h.ServeHTTP(w, r)
				return
			}

			if username == user && password == pass {
				r = r.WithContext(context.WithValue(r.Context(), IsAuthenticatedKey, true))
				r = r.WithContext(context.WithValue(r.Context(), ClaimsKey, username))
			}

			h.ServeHTTP(w, r)
		})
	}
}

func JWTBearerAuthenticationMiddleware[C any, CLAIMS interface {
	jwt.Claims
	*C
}](secret []byte) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			authorization, ok := strings.CutPrefix(authorization, "Bearer ")
			if !ok {
				h.ServeHTTP(w, r)
				return
			}
			claims := CLAIMS(new(C))
			tok, err := jwt.ParseWithClaims(authorization, claims, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				slog.Error("error in parsing jwt with claims", "err", err)
				h.ServeHTTP(w, r)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), ClaimsKey, tok.Claims))
			r = r.WithContext(context.WithValue(r.Context(), IsAuthenticatedKey, true))
			h.ServeHTTP(w, r)
		})
	}
}

func AuthenticatedOnlyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(IsAuthenticatedKey) != true {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized user"))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func ChainMiddlewares(ms ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		for i := len(ms) - 1; i >= 0; i-- {
			h = ms[i](h)
		}

		return h
	}
}
