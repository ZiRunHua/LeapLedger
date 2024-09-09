package v1

import (
	"KeepAccount/router/group"
)

func init() {
	// base path: /account/{accountId}
	router := group.Private.Group("account")
	baseApi := apiApp.AccountApi
	{
		router.POST("", baseApi.CreateOne)
		group.AccountAdministrator.PUT("", baseApi.Update)
		group.AccountAdministrator.DELETE("", baseApi.Delete)
		router.GET("/list", baseApi.GetList)
		router.GET("/list/:type", baseApi.GetListByType)
		group.Account.GET("", baseApi.GetOne)
		group.Account.GET("/info/:type", baseApi.GetInfo)
		group.Account.GET("/info", baseApi.GetInfo)
		// 模板
		router.GET("/template/list", baseApi.GetAccountTemplateList)
		router.POST("/form/template/:id", baseApi.CreateOneByTemplate)
		group.AccountCreator.POST("/transaction/category/init", baseApi.InitCategoryByTemplate)
		// 共享
		group.AccountCreator.PUT("/user/:id", baseApi.UpdateUser)
		group.Account.GET("/user/list", baseApi.GetUserList)
		group.Account.GET("/user/:id/info", baseApi.GetUserInfo)
		router.GET("/user/invitation/list", baseApi.GetUserInvitationList)
		group.AccountOwnEditor.POST("/user/invitation", baseApi.CreateAccountUserInvitation)
		router.PUT("/user/invitation/:id/accept", baseApi.AcceptAccountUserInvitation)
		router.PUT("/user/invitation/:id/refuse", baseApi.RefuseAccountUserInvitation)
		// 账本关联
		group.AccountOwnEditor.GET("/mapping", baseApi.GetAccountMapping)
		group.AccountOwnEditor.DELETE("/mapping/:id", baseApi.DeleteAccountMapping)
		group.Account.GET("/mapping/list", baseApi.GetAccountMappingList)
		group.AccountOwnEditor.POST("/mapping", baseApi.CreateAccountMapping)
		group.AccountOwnEditor.PUT("/mapping/:id", baseApi.UpdateAccountMapping)
		// 账本用户配置
		group.AccountOwnEditor.GET("/user/config", baseApi.GetUserConfig)
		group.AccountOwnEditor.PUT("/user/config/flag/:flag", baseApi.UpdateUserConfigFlag)
	}
}
