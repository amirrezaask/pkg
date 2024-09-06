package httpclient

import (
	"fmt"
	"net/http"
)

type mockTransport struct {
	requestToResponseErr map[string]struct {
		Response *http.Response
		Err      error
	}
}

func (m *mockTransport) AddRequest(method string, path string, resp *http.Response, err error) {
	m.requestToResponseErr[fmt.Sprintf("%s %s", method, path)] = struct {
		Response *http.Response
		Err      error
	}{resp, err}
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	method := req.Method
	path := req.URL.Path
	host := req.URL.Host
	s := m.requestToResponseErr[fmt.Sprintf("%s %s%s", method, host, path)]
	if s.Response == nil && s.Err == nil {
		s.Response = &http.Response{
			Status:     "OK 200",
			StatusCode: 200,
		}
	}
	return s.Response, s.Err
}

func NewHttpClientMock(target *http.Client) *mockTransport {
	mock := &mockTransport{}
	hc := http.Client{Transport: mock}
	*target = hc
	return mock
}
