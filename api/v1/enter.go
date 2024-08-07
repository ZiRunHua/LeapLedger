package v1

import (
	"KeepAccount/api/response"
	apiUtil "KeepAccount/api/util"
	"KeepAccount/global"
	"KeepAccount/service"
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
	gdb         = global.GvaDb
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
	//第三方服务
	thirdpartyService = service.GroupApp.ThirdpartyServiceGroup
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
