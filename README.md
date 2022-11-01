# tools
my simple tools, mostly golang code generators.

## enumgen
enumgen generates enum code. for example
if you have something like:
```go
//go:generate enumgen $GOFILE
package main

// enumgen: Started Arrived Finished
type RideState struct{}
```

and you run `go:generate`, you would get
```go

// GENERATED USING enum program, DONT EDIT BY HAND
package main

import "fmt"

type __ENUM__RideState struct {
	variant int
}

var (
	
	Started = __ENUM__RideState { 0 }
	
	Arrived = __ENUM__RideState { 1 }
	
	Finished = __ENUM__RideState { 2 }
	
)


func RideStateFromString(s string) (__ENUM__RideState, error) {
	switch s {
		
	case "Started":
		return __ENUM__RideState{ 0 }, nil
	
	case "Arrived":
		return __ENUM__RideState{ 1 }, nil
	
	case "Finished":
		return __ENUM__RideState{ 2 }, nil
	
	default:
		return __ENUM__RideState{}, fmt.Errorf("invalid RideState variant: %s", s)
	}
}


func (e __ENUM__RideState) String() string {
	switch e.variant {
		
	case 0:
		return "Started"
	
	case 1:
		return "Arrived"
	
	case 2:
		return "Finished"
	
	default:
		return ""
	} 
}

```

## sqlgen
sqlgen generates functions and structs to read and write from a sql database based on a struct.
for example if you have 
```go
//go:generate sqlgen $GOFILE
package main

// sqlgen: 
type User struct {
	Id        int
	Name, Age string
}
```
and if you now run `go:generate` you would get:
```go

package main

type UserWhereBuilder struct {
	
	Age *string
	
	Id *int
	
	Name *string
	
}


func (m *UserWhereBuilder) WhereAge(Age string) *UserWhereBuilder {
	m.Age = &Age
	return m
}

func (m *UserWhereBuilder) WhereId(Id int) *UserWhereBuilder {
	m.Id = &Id
	return m
}

func (m *UserWhereBuilder) WhereName(Name string) *UserWhereBuilder {
	m.Name = &Name
	return m
}


type UserQueryBuilder struct {
	UserWhereBuilder
}
	
func QueryUser() *UserQueryBuilder {
	return &UserQueryBuilder{}
}
	
type UserUpdateBuilder struct {
	set struct {
		
		Age *string
		
		Id *int
		
		Name *string
		
	}

	where UserWhereBuilder
}

func (m *UserUpdateBuilder) SetAge(Age string) *UserUpdateBuilder {
	m.set.Age = &Age
	return m
}

func (m *UserUpdateBuilder) SetId(Id int) *UserUpdateBuilder {
	m.set.Id = &Id
	return m
}

func (m *UserUpdateBuilder) SetName(Name string) *UserUpdateBuilder {
	m.set.Name = &Name
	return m
}

type UserDeleteBuilder struct {
	where UserWhereBuilder
}
```
