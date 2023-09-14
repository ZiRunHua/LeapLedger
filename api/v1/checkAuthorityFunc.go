package v1

import (
	"KeepAccount/api/v1/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	"github.com/gin-gonic/gin"
)

func checkAccountBelong(accountId interface{}, ctx *gin.Context) (bool, *accountModel.Account) {
	var account accountModel.Account
	err := global.GvaDb.First(&account, accountId).Error
	if err != nil {
		response.FailToError(ctx, err)
		return false, nil
	}
	if account.UserId != current.GetUserId(ctx) {
		response.Forbidden(ctx)
		return false, nil
	}
	return true, &account
}
