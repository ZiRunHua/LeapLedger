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
		publicRouter.GET("/captcha", publicApi.Captcha)
		publicRouter.POST("/captcha/email/send", publicApi.SendEmailCaptcha)

		publicRouter.POST("/user/login", publicApi.Login)
		publicRouter.POST("/user/register", publicApi.Register)
		publicRouter.PUT("/user/password", publicApi.UpdatePassword)
	}
	return publicRouter
}
