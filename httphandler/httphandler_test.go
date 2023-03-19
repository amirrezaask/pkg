package httphandler

import (
	"net/http"
	"testing"
)

func getUserHandlerStd(r *http.Request, body struct{}) {}

func TestStd(t *testing.T) {
	http.HandleFunc("/", StdHandler(getUserHandlerStd))
}
