package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// httpgen: handler user_create input
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// httpgen: handler user_create output
type CreateUserResponse struct {
	ID int `json:"id"`
}

// httpgen: ctx
type AppCtx struct {
	db *sql.DB
}

// generated
func MakeCreateUserHandler(h func(CreateUserRequest) (CreateUserResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input CreateUserRequest
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			panic(err)
		}

		resp, err := h(input)
		if err != nil {
			panic(err)
		}
		err = json.NewEncoder(w).Encode(resp)

		if err != nil {
			panic(err)
		}
	}
}
