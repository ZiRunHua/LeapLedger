package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type PublicRouter struct{}

func (s *PublicRouter) InitPublicRouter(_router *gin.RouterGroup) *gin.RouterGroup {
	router := _router.Group("public")
	publicApi := v1.ApiGroupApp.PublicApi
	{
		router.GET("/captcha", publicApi.Captcha)
		router.POST("/captcha/email/send", publicApi.SendEmailCaptcha)

		router.POST("/user/login", publicApi.Login)
		router.POST("/user/register", publicApi.Register)
		router.PUT("/user/password", publicApi.UpdatePassword)
		router.POST("/user/tour", publicApi.TourRequest)
	}
	return router
}
