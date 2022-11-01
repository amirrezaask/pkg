//go:generate sqlgen $GOFILE
package main

// sqlgen: table=users pk=id
type User struct {
	Id        int
	Name, Age string
}
