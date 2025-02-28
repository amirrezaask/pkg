package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestServerRouteRegisteration(t *testing.T) {
	type input struct{}
	type output struct{}
	tfs := []struct {
		Name string
		f    any
	}{
		{
			Name: "http form",
			f: func(r *Request) (Result, error) {
				return Result{Status: 200, Body: map[string]string{"Message": "Hello " + r.PathValue("name")}}, nil
			},
		},
		{
			Name: "std form",
			f: func(rw http.ResponseWriter, r *Request) {

			},
		},

		{
			Name: "reflect form",
			f: func(*Request, *input) (output, error) {
				return output{}, nil
			},
		},
	}
	mux := NewServeMux()
	for _, tf := range tfs {
		mux.HandleFunc("GET "+"/"+tf.Name, tf.f)
	}
}

func TestServerBind(t *testing.T) {
	is := is.New(t)
	type input struct {
		FromPath  int     `path:"from_path"`
		FromQuery float64 `query:"from_query"`
		FromBody  string  `json:"from_body"`
	}
	type output struct {
		FromPath  int     `json:"from_path"`
		FromQuery float64 `json:"from_query"`
		FromBody  string  `json:"from_body"`
	}
	tfs := []struct {
		Name string
		f    any
	}{
		{
			Name: "bind_reflect_form",
			f: func(r *Request, i *input) (output, error) {
				return output{
					FromPath:  i.FromPath,
					FromQuery: i.FromQuery,
					FromBody:  i.FromBody,
				}, nil
			},
		},
	}
	mux := NewServeMux()
	for _, tf := range tfs {
		mux.HandleFunc("POST "+"/{from_path}", tf.f)
	}

	srv := httptest.NewServer(mux)

	client := &http.Client{}
	bs, err := json.Marshal(map[string]interface{}{
		"from_body": "hello it's me again",
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", srv.URL+"/12?from_query=3.14", bytes.NewReader(bs))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	var o output
	err = json.NewDecoder(resp.Body).Decode(&o)
	if err != nil {
		t.Fatal(err)
	}
	is.Equal(o.FromPath, 12)
	is.Equal(o.FromQuery, 3.14)
	is.Equal(o.FromBody, "hello it's me again")
}
