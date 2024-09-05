package response

type Error struct {
	Message   string `json:"message,omitempty"`
	ErrorCode int    `json:"error_code,omitempty"`
}

type Success[T any] struct {
	Message string `json:"message,omitempty"`
	Data    T      `json:"data"`
}

var InternalServerError = Error{
	Message: "Internal Server Error",
}
