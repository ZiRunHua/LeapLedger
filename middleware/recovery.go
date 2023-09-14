package middleware

import (
	"KeepAccount/api/v1/response"
	"KeepAccount/global"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
)

// Recovery recover掉项目可能出现的panic

func Recovery(c *gin.Context, err any) {
	body, _ := io.ReadAll(c.Request.Body)

	global.PanicLogger.Error(
		"[Recovery from panic]",
		zap.Any("error", err),
		zap.String("method", c.Request.Method),
		zap.String("url", c.Request.RequestURI),
		zap.Any("body", body),
	)
	if str, ok := err.(string); ok {
		response.FailWithMessage(str, c)
	} else if e, ok := err.(error); ok {
		response.FailWithMessage(e.Error(), c)
	} else {
		response.Fail(c)
	}
}
