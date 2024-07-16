package v1

import (
	"KeepAccount/api/response"
	apiUtil "KeepAccount/api/util"
	"KeepAccount/global"
	userModel "KeepAccount/model/user"
	"KeepAccount/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	userService        = service.GroupApp.UserServiceGroup
	accountService     = service.GroupApp.AccountServiceGroup
	categoryService    = service.GroupApp.CategoryServiceGroup
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

func turnAwayTourist(user userModel.User, ctx *gin.Context, _db ...*gorm.DB) bool {
	var db *gorm.DB
	if len(_db) > 0 {
		db = _db[0]
	} else {
		db = global.GvaDb
	}
	isTourist, err := user.IsTourist(db)
	if responseError(err, ctx) {
		return true
	}
	if isTourist {
		response.FailToParameter(ctx, global.ErrTouristHaveNoRight)
		return true
	}
	return false
}
