package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type CategoryRouter struct{}

func (c *CategoryRouter) InitCategoryRouter(Router *gin.RouterGroup) {
	router := Router.Group("transaction/category")
	baseApi := v1.ApiGroupApp.CategoryApi
	{
		router.POST("", baseApi.CreateOne)
		router.POST("/:id/move", baseApi.MoveCategory)
		router.PUT("/:id", baseApi.Update)
		router.DELETE("/:id", baseApi.Delete)
		router.GET("/tree", baseApi.GetTree)
	}
	{
		router.POST("/father", baseApi.CreateOneFather)
		router.POST("/father/:id/move", baseApi.MoveFather)
		router.PUT("/father/:id", baseApi.UpdateFather)
		router.DELETE("/father/:id", baseApi.DeleteFather)
	}
	{
		router.POST("/:id/mapping", baseApi.MappingCategory)
		router.DELETE("/:id/mapping", baseApi.DeleteCategoryMapping)
		router.GET("/mapping/tree", baseApi.GetMappingTree)
	}
}
