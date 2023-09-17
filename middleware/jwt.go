package middleware

import (
	"KeepAccount/api/response"
	apiUtil "KeepAccount/api/util"
	"KeepAccount/util"
	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
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
