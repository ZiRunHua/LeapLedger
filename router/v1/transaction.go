package v1

import (
	v1 "KeepAccount/api/v1"
	"github.com/gin-gonic/gin"
)

type TransactionRouter struct{}

func (c *TransactionRouter) InitTransactionRouter(Router *gin.RouterGroup) {
	router := Router.Group("transaction")
	baseApi := v1.ApiGroupApp.TransactionApi
	{
		router.GET("/:id", baseApi.GetOne)
		router.POST("", baseApi.CreateOne)
		router.PUT("/:id", baseApi.Update)
		router.DELETE("/:id", baseApi.Delete)
		router.GET("/list", baseApi.GetList)
		router.GET("/total", baseApi.GetTotal)
		router.GET("/month/statistic", baseApi.GetMonthStatistic)
		router.GET("/day/statistic", baseApi.GetDayStatistic)
		router.GET("/category/amount/rank", baseApi.GetCategoryAmountRank)
	}
}
