package v1

import (
	"github.com/ZiRunHua/LeapLedger/api/response"
	apiUtil "github.com/ZiRunHua/LeapLedger/api/util"
	"github.com/ZiRunHua/LeapLedger/service"
	"github.com/gin-gonic/gin"
)

type PublicApi struct {
}

type ApiGroup struct {
	AccountApi
	CategoryApi
	UserApi
	TransactionApi
	PublicApi
	ProductApi
}

var (
	ApiGroupApp = new(ApiGroup)
)

// 服务
var (
	commonService = service.GroupApp.CommonServiceGroup
)
var (
	userService        = service.GroupApp.UserServiceGroup
	accountService     = service.GroupApp.AccountServiceGroup
	categoryService    = service.GroupApp.CategoryServiceGroup
	transactionService = service.GroupApp.TransactionServiceGroup
	productService     = service.GroupApp.ProductServiceGroup
	templateService    = service.GroupApp.TemplateServiceGroup
)

// 工具
var contextFunc = apiUtil.ContextFunc
var checkFunc = apiUtil.CheckFunc

func handelError(err error, ctx *gin.Context) bool {
	if err != nil {
		response.FailToError(ctx, err)
		return true
	}
	return false
}

func responseError(err error, ctx *gin.Context) bool {
	if err != nil {
		response.FailToError(ctx, err)
		return true
	}
	return false
}
