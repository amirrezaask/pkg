//go:generate sqlgen $GOFILE
package main

// sqlgen:
type User struct {
	Id   int
	Name string
	Age  int
}
