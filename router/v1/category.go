package v1

import (
	"github.com/ZiRunHua/LeapLedger/router/group"
)

func init() {
	// base path: /account/{accountId}/category
	readRouter := group.Account.Group("category")
	editRouter := group.AccountCreator.Group("category")
	editMappingRouter := group.AccountOwnEditor.Group("category")
	baseApi := apiApp.CategoryApi
	{
		editRouter.POST("", baseApi.CreateOne)
		editRouter.PUT("/:id/move", baseApi.MoveCategory)
		editRouter.PUT("/:id", baseApi.Update)
		editRouter.DELETE("/:id", baseApi.Delete)
		readRouter.GET("/tree", baseApi.GetTree)
		readRouter.GET("/list", baseApi.GetList)
	}
	{
		editRouter.POST("/father", baseApi.CreateOneFather)
		editRouter.PUT("/father/:id/move", baseApi.MoveFather)
		editRouter.PUT("/father/:id", baseApi.UpdateFather)
		editRouter.DELETE("/father/:id", baseApi.DeleteFather)
	}
	{
		editMappingRouter.POST("/:id/mapping", baseApi.MappingCategory)
		editMappingRouter.DELETE("/:id/mapping", baseApi.DeleteCategoryMapping)
		editMappingRouter.GET("/mapping/tree", baseApi.GetMappingTree)
	}
}
