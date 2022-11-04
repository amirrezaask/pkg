//go:generate httpgen example.go
package main

import (
	"database/sql"
)

type AppCtx struct {
	db *sql.DB
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	ID int `json:"id"`
}

// handler
func createUser(appCtx *AppCtx, req CreateUserRequest) (CreateUserResponse, error) {
	return CreateUserResponse{}, nil
}
