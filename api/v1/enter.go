package v1

import (
	"KeepAccount/api/response"
	apiUtil "KeepAccount/api/util"
	"KeepAccount/service"
	"github.com/gin-gonic/gin"
)

// 接口
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

var ApiGroupApp = new(ApiGroup)

// 服务
var (
	commonService = service.GroupApp.CommonServiceGroup
)
var (
	userService        = service.GroupApp.UserServiceGroup.User
	accountService     = service.GroupApp.AccountServiceGroup
	categoryService    = service.GroupApp.CategoryServiceGroup.Category
	transactionService = service.GroupApp.TransactionServiceGroup.Transaction
	productService     = service.GroupApp.ProductServiceGroup.Product
	templateService    = service.GroupApp.TemplateService.Template
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
