package v1

import (
	"KeepAccount/api/response"
	apiUtil "KeepAccount/api/util"
	"KeepAccount/service"
	commonService "KeepAccount/service/common"
	"github.com/gin-gonic/gin"
)

// api
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
	common = commonService.Common
)
var (
	userService        = service.GroupApp.UserServiceGroup.User
	accountService     = service.GroupApp.AccountServiceGroup
	categoryService    = service.GroupApp.CategoryServiceGroup.Category
	transactionService = service.GroupApp.TransactionServiceGroup.Transaction
	productService     = service.GroupApp.ProductServiceGroup.Product
)

// util
var contextFunc = apiUtil.ContextFunc
var checkFunc = apiUtil.CheckFunc

func handelError(err error, ctx *gin.Context) bool {
	if err != nil {
		response.FailToError(ctx, err)
		return true
	}
	return false
}
