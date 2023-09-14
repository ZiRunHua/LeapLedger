package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type TransactionImportRouter struct{}

func (a *TransactionImportRouter) InitTransactionImportRouter(Router *gin.RouterGroup) {
	router := Router.Group("transaction/import")
	baseApi := v1.ApiGroupApp.ProductApi
	{
		router.GET("/product/list", baseApi.GetList)
		router.GET("/product/:key/transaction/category", baseApi.GetTransactionCategory)
		router.GET("/product/category/mapping/tree", baseApi.GetMappingTree)
		router.GET("/product/transaction/category/:id/mapping", baseApi.MappingTransactionCategory)
		router.POST("/product/:key/import/bill", baseApi.ImportProductBill)
	}
}
