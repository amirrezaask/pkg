package httphandler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func getUserHandlerGin(c *gin.Context, body User) {
	return
}

func TestGin(t *testing.T) {
	g := gin.New()
	g.GET("/", GinHandler(getUserHandlerGin))
}
