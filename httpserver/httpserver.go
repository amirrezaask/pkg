package httpserver

import (
	"net/http"
)

type ServeMux struct {
	*http.ServeMux
	middlewares []Middleware
}

func New() *ServeMux {
	return &ServeMux{ServeMux: http.NewServeMux()}
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func (s *ServeMux) WithMiddlewares(ms ...Middleware) {
	s.middlewares = append(s.middlewares, ms...)
}

func (s *ServeMux) Group(path string, f func(innerServer *ServeMux), ms ...Middleware) {
	server := New()
	server.WithMiddlewares(s.middlewares...)
	server.WithMiddlewares(ms...)
	f(server)
	s.Handle(path, server)
}

func (s *ServeMux) HandleFunc(path string, handler http.HandlerFunc) {
	for _, m := range s.middlewares {
		handler = m(handler)
	}

	s.HandleFunc(path, handler)
}

func (s *ServeMux) GET(path string, handler http.HandlerFunc) {
	s.HandleFunc("GET "+path, handler)
}

func (s *ServeMux) POST(path string, handler http.HandlerFunc) {
	s.HandleFunc("POST "+path, handler)
}

func (s *ServeMux) PUT(path string, handler http.HandlerFunc) {
	s.HandleFunc("PUT "+path, handler)
}

func (s *ServeMux) PATCH(path string, handler http.HandlerFunc) {
	s.HandleFunc("PATCH "+path, handler)
}

func (s *ServeMux) DELETE(path string, handler http.HandlerFunc) {
	s.HandleFunc("DELETE "+path, handler)
}

func (s *ServeMux) OPTION(path string, handler http.HandlerFunc) {
	s.HandleFunc("OPTIONS "+path, handler)
}
