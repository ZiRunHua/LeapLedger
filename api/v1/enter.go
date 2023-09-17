package v1

import (
	"KeepAccount/api/response"
	"KeepAccount/service"
	commonService "KeepAccount/service/common"
	"github.com/gin-gonic/gin"
)

type PublicApi struct {
}

type ApiGroup struct {
	AccountApi
	CategoryApi
	UserApi
	PublicApi
	ProductApi
}

var ApiGroupApp = new(ApiGroup)

// service
var (
	common  = commonService.Common
	current = commonService.Current
)
var (
	userService        = service.GroupApp.UserServiceGroup.User
	accountService     = service.GroupApp.AccountServiceGroup.Account
	categoryService    = service.GroupApp.CategoryServiceGroup.Category
	transactionService = service.GroupApp.TransactionServiceGroup.Transaction
	productService     = service.GroupApp.ProductServiceGroup.Product
)

func handelError(err error, ctx *gin.Context) bool {
	if err != nil {
		response.FailToError(ctx, err)
		return true
	}
	return false
}
