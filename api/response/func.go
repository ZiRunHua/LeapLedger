package response

import (
	"KeepAccount/global"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Data struct {
	Data interface{}
	Msg  string `example:"success"`
} // @name Response

type NoContent struct {
	Data interface{}
	Msg  string
} // @name NoContent

func ResponseAndAbort(status int, data interface{}, msg string, ctx *gin.Context) {
	ctx.AbortWithStatusJSON(
		status, Data{
			data,
			msg,
		},
	)
}

func Response(status int, data interface{}, msg string, ctx *gin.Context) {
	ctx.JSON(
		status, Data{
			data,
			msg,
		},
	)
}

func Ok(ctx *gin.Context) {
	Response(204, map[string]interface{}{}, "操作成功", ctx)
}

func OkWithMessage(message string, ctx *gin.Context) {
	Response(200, map[string]interface{}{}, message, ctx)
}

func OkWithData(data interface{}, ctx *gin.Context) {
	Response(200, data, "查询成功", ctx)
}

func OkWithDetailed(data interface{}, message string, ctx *gin.Context) {
	Response(200, data, message, ctx)
}

func Fail(ctx *gin.Context) {
	ResponseAndAbort(500, map[string]interface{}{}, "服务器睡了（这年龄你睡得着！）", ctx)
}
func FailToParameter(ctx *gin.Context, err error) {
	ResponseAndAbort(400, map[string]interface{}{}, "参数错误"+err.Error(), ctx)
}

func FailToError(ctx *gin.Context, err error) {
	logError(ctx, err)
	msg := err.Error()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		msg = "数据未找到"
	}
	ResponseAndAbort(500, map[string]interface{}{}, msg, ctx)
}

func FailWithMessage(message string, ctx *gin.Context) {
	ResponseAndAbort(500, map[string]interface{}{}, message, ctx)
}

func FrequentOperation(ctx *gin.Context) {
	ResponseAndAbort(500, map[string]interface{}{}, "请勿频繁操作，请稍后再试！", ctx)
}

func FailWithDetailed(data interface{}, message string, ctx *gin.Context) {
	ResponseAndAbort(500, data, message, ctx)
}

func Forbidden(ctx *gin.Context) {
	ResponseAndAbort(403, map[string]interface{}{}, "无权访问", ctx)
}

func TokenExpired(ctx *gin.Context) {
	ResponseAndAbort(401, map[string]interface{}{}, "token expired", ctx)
}

func logError(ctx *gin.Context, err error) {
	reqMethod := ctx.Request.Method
	reqPath := ctx.Request.URL.Path
	clientIP := ctx.ClientIP()
	global.ErrorLogger.Error(
		fmt.Sprintf(
			"%s %s %s error: %+v\n",
			reqMethod,
			reqPath,
			clientIP,
			err,
		),
	)
}
