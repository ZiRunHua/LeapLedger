package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type PublicRouter struct{}

func (s *PublicRouter) InitPublicRouter(Router *gin.RouterGroup) *gin.RouterGroup {
	publicRouter := Router.Group("public")
	publicApi := v1.ApiGroupApp.PublicApi
	{
		publicRouter.POST("captcha", publicApi.Captcha)
		publicRouter.POST("login", publicApi.Login)

	}
	return publicRouter
}
