package v1

import (
	"github.com/ZiRunHua/LeapLedger/router/group"
)

func init() {
	// base path: /public
	router := group.Public
	{
		router.GET("/captcha", publicApi.Captcha)
		router.POST("/captcha/email/send", publicApi.SendEmailCaptcha)
	}
}
