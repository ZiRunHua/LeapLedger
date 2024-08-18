package v1

import (
	"KeepAccount/router/group"
)

func init() {
	// base path: /product
	router := group.Private.Group("product")
	// base path: /account/{accountId}/product
	accountRouter := group.Account.Group("product")
	editRouter := group.AccountCreator.Group("product")
	baseApi := apiApp.ProductApi
	{
		router.GET("/list", baseApi.GetList)
		router.GET("/:key/transCategory", baseApi.GetTransactionCategory)
		accountRouter.GET("/transCategory/mapping/tree", baseApi.GetMappingTree)
		editRouter.POST("/transCategory/:id/mapping", baseApi.MappingTransactionCategory)
		editRouter.DELETE("/transCategory/:id/mapping", baseApi.DeleteTransactionCategoryMapping)
		editRouter.POST("/:key/bill/import", baseApi.ImportProductBill)
	}
}
