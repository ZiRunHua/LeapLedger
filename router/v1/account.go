package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type AccountRouter struct{}

func (a *AccountRouter) InitAccountRouter(Router *gin.RouterGroup) {
	router := Router.Group("account")
	baseApi := v1.ApiGroupApp.AccountApi
	{
		router.POST("", baseApi.CreateOne)
		router.PUT("/:id", baseApi.Update)
		router.DELETE("/:id", baseApi.Delete)
		router.GET("/list", baseApi.GetList)
		router.GET("/list/:type", baseApi.GetListByType)
		router.GET("/:id", baseApi.GetOne)
		router.GET("/:id/info/:type", baseApi.GetInfo)
		router.GET("/:id/info", baseApi.GetInfo)
		//模板
		router.GET("/template/list", baseApi.GetAccountTemplateList)
		router.POST("/form/template/:id", baseApi.CreateOneByTemplate)
		router.POST("/:id/transaction/category/init", baseApi.InitTransCategoryByTemplate)
		//共享
		router.PUT("/user/:id", baseApi.UpdateUser)
		router.GET("/:id/user/list", baseApi.GetUserList)
		router.GET("/user/:id/info", baseApi.GetUserInfo)
		router.GET("/user/invitation/list", baseApi.GetUserInvitationList)
		router.POST("/:id/user/invitation", baseApi.CreateAccountUserInvitation)
		router.POST("/user/invitation/:id/accept", baseApi.AcceptAccountUserInvitation)
		router.POST("/user/invitation/:id/refuse", baseApi.RefuseAccountUserInvitation)
		//账本关联
		router.GET("/:id/mapping", baseApi.GetAccountMapping)
		router.GET("/:id/mapping/list", baseApi.GetAccountMappingList)
		router.DELETE("/mapping/:id", baseApi.DeleteAccountMapping)
		router.POST("/:id/mapping", baseApi.CreateAccountMapping)
		router.PUT("/:id/mapping", baseApi.UpdateAccountMapping)
	}
}
