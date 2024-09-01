package v1

import (
	"KeepAccount/router/group"
)

func init() {
	// base path: /user or /public/user
	router := group.Private.Group("user")
	publicRouter := group.Public.Group("user")
	noTouristRouter := group.NoTourist.Group("user")
	baseApi := apiApp.UserApi
	{
		// public
		publicRouter.POST("/login", publicApi.Login)
		publicRouter.POST("/register", publicApi.Register)
		publicRouter.PUT("/password", publicApi.UpdatePassword)
		publicRouter.POST("/tour", publicApi.TourRequest)
		// current user
		noTouristRouter.POST("/current/captcha/email/send", baseApi.SendCaptchaEmail)
		router.POST("/token/refresh", baseApi.RefreshToken)
		router.PUT("/client/current/account", baseApi.SetCurrentAccount)
		router.PUT("/client/current/share/account", baseApi.SetCurrentShareAccount)
		noTouristRouter.PUT("/current/password", baseApi.UpdatePassword)
		noTouristRouter.PUT("/current", baseApi.UpdateInfo)
		router.GET("/home", baseApi.Home)
		// all user
		router.GET("/search", baseApi.SearchUser)
		// config
		router.GET("/transaction/share/config", baseApi.GetTransactionShareConfig)
		router.PUT("/transaction/share/config", baseApi.UpdateTransactionShareConfig)
		// friend
		router.GET("/friend/list", baseApi.GetFriendList)
		router.POST("/friend/invitation", baseApi.CreateFriendInvitation)
		router.PUT("/friend/invitation/:id/accept", baseApi.AcceptFriendInvitation)
		router.PUT("/friend/invitation/:id/refuse", baseApi.RefuseFriendInvitation)
		router.GET("/friend/invitation", baseApi.GetFriendInvitationList)

		router.GET("/account/invitation/list", baseApi.GetAccountInvitationList)
	}
}
