package httphandler

import (
	"fmt"
	"net/http"
	"testing"
)

// func getUser(w http.ResponseWriter, r *http.Request) {
// 	var req User
// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		w.WriteHeader(400)
// 		return
// 	}

//		fmt.Println(req.Name)
//	}
func getUserHandlerStd(w http.ResponseWriter, r *http.Request, body User) {
	fmt.Println(body.Name)
}

func TestStd(t *testing.T) {
	// StdCallbacks.BodyDecoder = func(req *http.Request, out any) error {
	// 	return xml.NewDecoder(req.Body).Decode(out)
	// }
	http.HandleFunc("/", StdHandler(getUserHandlerStd))
}
