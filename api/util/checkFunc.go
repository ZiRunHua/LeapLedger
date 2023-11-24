package util

import (
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	"github.com/gin-gonic/gin"
)

type _checkFunc interface {
	AccountBelong(accountId interface{}, ctx *gin.Context) (bool, *accountModel.Account)
	TransactionCategoryBelong(id interface{}, ctx *gin.Context) (pass bool, category *categoryModel.Category, account *accountModel.Account)
}

type checkFunc struct {
}

var CheckFunc = new(checkFunc)

func (ckf *checkFunc) AccountBelong(accountId interface{}, ctx *gin.Context) (bool, *accountModel.Account) {
	var account accountModel.Account
	err := global.GvaDb.First(&account, accountId).Error
	if err != nil {
		response.FailToError(ctx, err)
		return false, nil
	}
	if account.UserId != ContextFunc.GetUserId(ctx) {
		response.Forbidden(ctx)
		return false, nil
	}
	return true, &account
}

func (ckf *checkFunc) TransactionCategoryBelong(id interface{}, ctx *gin.Context) (pass bool, category *categoryModel.Category, account *accountModel.Account) {
	err := global.GvaDb.First(&category, id).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	account, err = category.GetAccount()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if account.UserId != ContextFunc.GetUserId(ctx) {
		response.Forbidden(ctx)
		return
	}
	return true, category, account
}
