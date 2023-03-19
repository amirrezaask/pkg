package httphandler

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

var GinCallbacks = Callbacks[*gin.Context]{
	BodyDecoder: func(req *gin.Context, out any) error {
		return json.NewDecoder(req.Request.Body).Decode(out)
	},
}

func GinHandler[IN any](handler func(c *gin.Context, body IN)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req IN
		_ = GinCallbacks.BodyDecoder(c, req)
		handler(c, req)
	}
}
