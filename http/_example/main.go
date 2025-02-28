package main

import (
	"log/slog"
	"os"

	"github.com/amirrezaask/pkg/http"
	"github.com/golang-jwt/jwt/v5"
)

func main() {

	type claims struct {
		jwt.Claims
	}
	mux := http.NewServeMux()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	mux.UseMiddlewares(
		http.PrometheusExporterMiddleware("App"),
		http.JWTBearerAuthenticationMiddleware[claims]([]byte("")),
		http.RequestLoggerMiddleware(os.Stdout),
		http.RecoverMiddleware,
	)
	type createUserRequest struct {
		Name string
	}
	type createUserResponse struct {
		Message string
	}

	mux.HandleFunc("GET /std/{name}", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Hello " + r.PathValue("name")))
		rw.WriteHeader(200)
	})

	mux.HandleFunc("GET /simple/{name}", func(r *http.Request) (http.Result, error) {
		return http.Result{Status: 200, Body: map[string]string{"Message": "Hello " + r.PathValue("name")}}, nil
	})

	mux.HandleFunc("GET /reflect/{name}", func(r *http.Request, req *createUserRequest) (createUserResponse, error) {
		return createUserResponse{Message: "Hello World " + r.PathValue("name")}, nil
	})

	mux.HandleFunc("POST /reflect/post/{id}", func(r *http.Request, req *struct {
		Name string `json:"name"`
		ID   int64  `path:"id"`
	}) (output struct {
		Name string `json:"name"`
		ID   int64  `json:"id"`
	}, err error) {
		output.Name = req.Name
		output.ID = req.ID
		return output, nil
	})

	mux.MapPrometheusEndpoint("/metrics")
	http.ListenAndServe("localhost:8080", mux)
}
