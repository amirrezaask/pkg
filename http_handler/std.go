package httphandler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var StdErrorEncoder func(err any) ([]byte, error) = json.Marshal
var StdRespEncoder func(t any) ([]byte, error) = json.Marshal
var StdReqDecoder func(r io.Reader, t any) error = func(r io.Reader, t any) error {
	return json.NewDecoder(r).Decode(t)
}

type Request[T any] struct {
	Body    T
	Headers http.Header
	Query   url.Values
}
type Response[T any] struct {
	Body   T
	Status int
}

func MakeHTTPHandler[IN, OUT any](handler func(Request[IN]) (Response[OUT], error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req IN
		_ = StdReqDecoder(r.Body, req)
		resp, err := handler(Request[IN]{Body: req, Headers: r.Header, Query: r.URL.Query()})
		w.WriteHeader(resp.Status)
		if err != nil {
			bs, _ := StdErrorEncoder(err)
			fmt.Fprint(w, bs)
			return
		}
		bs, _ := StdRespEncoder(resp.Body)
		fmt.Fprint(w, bs)
	}
}

// Sample
// type User struct{}

// func getUser(r Request[User]) (Response[User], error) {

// 	return Response[User]{
// 		Body:   User{},
// 		Status: 201,
// 	}, nil
// }

// func main() {
// 	http.HandleFunc("", MakeHTTPHandler(getUser))
// }
