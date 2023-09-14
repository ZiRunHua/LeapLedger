package middleware

import (
	"KeepAccount/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Request.Header.Get("authorization")
		if token == "" {
			ctx.JSON(
				http.StatusUnauthorized, gin.H{
					"msg":  "未登录或非法访问",
					"data": gin.H{},
				},
			)
			ctx.Abort()
			return
		}
		jwt := util.NewJWT()
		// parseToken 解析token包含的信息
		claims, err := jwt.ParseToken(token)
		if err != nil {
			if err == util.TokenExpired {
				ctx.JSON(
					http.StatusUnauthorized, gin.H{
						"data": gin.H{},
						"msg":  "授权已过期",
					},
				)
				ctx.Abort()
				return
			}
			ctx.JSON(
				http.StatusUnauthorized, gin.H{
					"data": gin.H{},
					"msg":  err.Error(),
				},
			)
			ctx.Abort()
			return
		}
		ctx.Set("claims", claims)
		ctx.Set("userId", claims.UserId)
		ctx.Next()
	}
}
