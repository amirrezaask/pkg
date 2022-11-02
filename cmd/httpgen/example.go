package main

// httpgen: user_create input
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// httpgen: user_create output
type CreateUserResponse struct {
	ID int `json:"id"`
}
