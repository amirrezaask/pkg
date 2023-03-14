# Pkg

## Http Handler

make http handling better.

### Standard library http

```go
package main

import (
    "github.com/amirrezaask/pkg/http_handler"
)

type User struct {}

func getUserHandler(r *http.Request) (User, error) {}

func main() {
    http.HandleFunc("/user", httphandler.StdHTTP(getUserHandler))
}
```

### Echo

```go
package main

import (
    "github.com/amirrezaask/pkg/http_handler"
)

type User struct {}

func getUserHandler(c echo.Context) (User, error) {}

func main() {
    e := echo.New()
    e.GET("/user", httphandler.Echo(getUserHandler))
}
```
