//go:generate qbgen -file $GOFILE -dialect mysql
package main

//qbgen: model
type User struct {
	ID       int64
	Username string
	Password string
}
