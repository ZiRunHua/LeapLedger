package router

import (
	"KeepAccount/api/response"
	apiUtil "KeepAccount/api/util"
	"KeepAccount/global"
	"KeepAccount/util"
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"time"
)

type _middleware struct {
}

var middleware = &_middleware{}

func (m *_middleware) JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := apiUtil.ContextFunc.GetToken(ctx)
		if token == "" {
			response.Forbidden(ctx)
			return
		}
		jwt := util.NewJWT()
		// parseToken 解析token包含的信息
		claims, err := jwt.ParseToken(token)
		if err != nil {
			if err == util.TokenExpired {
				response.TokenExpired(ctx)
				return
			}
			response.FailToError(ctx, err)
			return
		}
		apiUtil.ContextFunc.SetUserId(claims.UserId, ctx)
		apiUtil.ContextFunc.SetClaims(claims, ctx)
		ctx.Next()
	}
}

func (m *_middleware) Recovery(ctx *gin.Context, err any) {
	body, _ := io.ReadAll(ctx.Request.Body)
	global.PanicLogger.Error(
		"[Recovery from panic]",
		zap.Any("error", err),
		zap.String("method", ctx.Request.Method),
		zap.String("url", ctx.Request.RequestURI),
		zap.Any("body", body),
	)
	response.Fail(ctx)
}

func (m *_middleware) RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		bodyBytes, err := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
		var body string
		if err != nil {
			body = ""
		} else {
			body = string(bodyBytes)
		}
		cost := time.Since(start)
		logger.Info(
			path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("body", body),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}
