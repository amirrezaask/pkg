package pkg

import (
	"testing"

	"github.com/labstack/echo/v4"
)

type UserInfo struct{}

func TestCtx(t *testing.T) {
	var c echo.Context

	// obj := c.Request().Context().Value("userinfo")
	// userInfo, ok := obj.(UserInfo)
	// if !ok {
	// 	c.NoContent(400)
	// }

	// _ = userInfo

	userInfo, err := CtxGetValue[UserInfo](c.Request().Context(), "userinfo")

	_ = err
	_ = userInfo
}
