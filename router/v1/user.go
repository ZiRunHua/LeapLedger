package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (s *PublicRouter) InitUserRouter(_router *gin.RouterGroup, _turnAwayTouristRouter *gin.RouterGroup) {
	router := _router.Group("user")
	turnAwayTouristRouter := _turnAwayTouristRouter.Group("user")
	baseApi := v1.ApiGroupApp.UserApi
	{
		router.POST("/token/refresh", baseApi.RefreshToken)

		router.GET("/search", baseApi.SearchUser)
		turnAwayTouristRouter.POST("/current/captcha/email/send", baseApi.SendCaptchaEmail)
		router.PUT("/client/current/account", baseApi.SetCurrentAccount)
		router.PUT("/client/current/share/account", baseApi.SetCurrentShareAccount)
		turnAwayTouristRouter.PUT("/current/password", baseApi.UpdatePassword)
		router.PUT("/current", baseApi.UpdateInfo)
		router.GET("/home", baseApi.Home)

		router.GET("/transaction/share/config", baseApi.GetTransactionShareConfig)
		router.PUT("/transaction/share/config", baseApi.UpdateTransactionShareConfig)

		router.GET("/friend/list", baseApi.GetFriendList)
		router.POST("/friend/invitation", baseApi.CreateFriendInvitation)
		router.POST("/friend/invitation/:id/accept", baseApi.AcceptFriendInvitation)
		router.POST("/friend/invitation/:id/refuse", baseApi.RefuseFriendInvitation)
		router.GET("/friend/invitation", baseApi.GetFriendInvitationList)

		router.GET("/account/invitation/list", baseApi.GetAccountInvitationList)
	}
}
