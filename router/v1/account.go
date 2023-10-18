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
		router.GET("/:id", baseApi.GetOne)
		//模板
		router.GET("/template/list", baseApi.GetAccountTemplateList)
		router.POST("/form/template/:id", baseApi.CreateOneByTemplate)

	}
}
