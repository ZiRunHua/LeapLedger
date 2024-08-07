package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type AccountRouter struct{}

func (a *AccountRouter) InitAccountRouter(_router *gin.RouterGroup) {
	router := _router.Group("account")
	baseApi := v1.ApiGroupApp.AccountApi
	{
		router.POST("", baseApi.CreateOne)
		router.PUT("/:accountId", baseApi.Update)
		router.DELETE("/:accountId", baseApi.Delete)
		router.GET("/list", baseApi.GetList)
		router.GET("/list/:type", baseApi.GetListByType)
		router.GET("/:accountId", baseApi.GetOne)
		router.GET("/:accountId/info/:type", baseApi.GetInfo)
		router.GET("/:accountId/info", baseApi.GetInfo)
		//模板
		router.GET("/template/list", baseApi.GetAccountTemplateList)
		router.POST("/form/template/:id", baseApi.CreateOneByTemplate)
		router.POST("/:accountId/transaction/category/init", baseApi.InitCategoryByTemplate)
		//共享
		router.PUT("/user/:id", baseApi.UpdateUser)
		router.GET("/:accountId/user/list", baseApi.GetUserList)
		router.GET("/user/:id/info", baseApi.GetUserInfo)
		router.GET("/user/invitation/list", baseApi.GetUserInvitationList)
		router.POST("/:accountId/user/invitation", baseApi.CreateAccountUserInvitation)
		router.POST("/user/invitation/:id/accept", baseApi.AcceptAccountUserInvitation)
		router.POST("/user/invitation/:id/refuse", baseApi.RefuseAccountUserInvitation)
		//账本关联
		router.GET("/:accountId/mapping", baseApi.GetAccountMapping)
		router.GET("/:accountId/mapping/list", baseApi.GetAccountMappingList)
		router.DELETE("/mapping/:id", baseApi.DeleteAccountMapping)
		router.POST("/:accountId/mapping", baseApi.CreateAccountMapping)
		router.PUT("/mapping/:id", baseApi.UpdateAccountMapping)
		//账本用户配置
		router.GET("/user/config", baseApi.GetUserConfig)
		router.PUT("/user/config/flag/:type", baseApi.UpdateUserConfigFlag)
	}
}
