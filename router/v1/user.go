package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (s *PublicRouter) InitUserRouter(Router *gin.RouterGroup) {
	router := Router.Group("user")
	baseApi := v1.ApiGroupApp.UserApi
	{
		router.POST("/current/captcha/email/send", baseApi.SendCaptchaEmail)
		router.PUT("/client/current/account", baseApi.SetCurrentAccount)
		router.PUT("/current/password", baseApi.UpdatePassword)
		router.PUT("/current", baseApi.UpdateInfo)
	}
}
