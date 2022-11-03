package main

import (
	"database/sql"
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
