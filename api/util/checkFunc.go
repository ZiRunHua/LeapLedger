package util

import (
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	"github.com/gin-gonic/gin"
)

type _checkFunc interface {
	AccountBelong(accountId interface{}, ctx *gin.Context) (bool, *accountModel.Account)
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
