package main

import (
	"database/sql"
)

// httpgen: input UserCreate
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// httpgen: output UserCreate
type CreateUserResponse struct {
	ID int `json:"id"`
}

// httpgen: ctx
type AppCtx struct {
	db *sql.DB
}
