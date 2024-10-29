package v1

import (
	"github.com/ZiRunHua/LeapLedger/router/group"
)

func init() {
	// base path: /account/{accountId}/transaction
	readRouter := group.AccountReader.Group("transaction")
	editRouter := group.AccountOwnEditor.Group("transaction")
	baseApi := apiApp.TransactionApi
	{
		readRouter.GET("/:id", baseApi.GetOne)
		editRouter.POST("", baseApi.CreateOne)
		editRouter.PUT("/:id", baseApi.Update)
		editRouter.DELETE("/:id", baseApi.Delete)

		readRouter.GET("/list", baseApi.GetList)
		readRouter.GET("/total", baseApi.GetTotal)
		readRouter.GET("/month/statistic", baseApi.GetMonthStatistic)
		readRouter.GET("/day/statistic", baseApi.GetDayStatistic)
		readRouter.GET("/category/amount/rank", baseApi.GetCategoryAmountRank)
		readRouter.GET("/amount/rank", baseApi.GetAmountRank)
		// timing
		editRouter.GET("/timing/list", baseApi.GetTimingList)
		editRouter.POST("/timing", baseApi.CreateTiming)
		editRouter.PUT("/timing/:id", baseApi.UpdateTiming)
		editRouter.DELETE("/timing/:id", baseApi.DeleteTiming)
		editRouter.PUT("/timing/:id/:operate", baseApi.HandleTiming)
	}
}
