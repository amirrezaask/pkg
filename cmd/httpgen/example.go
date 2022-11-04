//go:generate httpgen example.go
package main

import (
	"database/sql"
)

// httpgen: ctx
type AppCtx struct {
	db *sql.DB
}

// httpgen: input UserCreate
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// httpgen: output UserCreate
type CreateUserResponse struct {
	ID int `json:"id"`
}

// httpgen: handler UserCreate
func createUser(appCtx *AppCtx, req CreateUserRequest) (CreateUserResponse, error) {
	return CreateUserResponse{}, nil
}
