package v1

import (
	v1 "KeepAccount/api/v1"
	"KeepAccount/global/cusCtx"
	"github.com/gin-gonic/gin"
)

type TransactionRouter struct{}

func (c *TransactionRouter) InitTransactionRouter(sRouter *gin.RouterGroup, accountAccountRoute *gin.RouterGroup) {
	router := sRouter.Group("transaction")
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
		router.GET("/amount/rank", baseApi.GetAmountRank)
		// timing
		accountAccountRoute.GET("/transaction/timing/list", baseApi.GetTimingList)
		accountAccountRoute.POST("/transaction/timing", baseApi.CreateTiming)
		accountAccountRoute.PUT("/transaction/timing/:"+string(cusCtx.TransactionTimingId), baseApi.UpdateTiming)
	}
}
