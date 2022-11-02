package main

// repogen: model user
type User struct{}

// repogen: repo user
type UserRepository interface {
	FindByID()
	FindByUserNameAndID()
	UpdateUserName()
}
