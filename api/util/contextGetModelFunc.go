package util

import (
	"KeepAccount/api/response"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"github.com/gin-gonic/gin"
)

func (cf *contextFunc) GetTransByParam(ctx *gin.Context) (result transactionModel.Transaction, pass bool) {
	id, ok := cf.GetParamId(ctx)
	if false == ok {
		return
	}
	trans := transactionModel.Transaction{}
	if err := trans.SelectById(id); err != nil {
		response.FailToError(ctx, err)
		return
	}
	if pass = CheckFunc.AccountBelong(trans.AccountId, ctx); false == pass {
		return
	}
	return trans, true
}

// GetAccountByParam 返回pass表示是否获取成功
func (cf *contextFunc) GetAccountByParam(ctx *gin.Context, checkBelong bool) (
	account accountModel.Account, accountUser accountModel.User, pass bool,
) {
	id, ok := cf.GetAccountIdByParam(ctx)
	if false == ok {
		return
	}
	if checkBelong {
		if account, accountUser, pass = CheckFunc.AccountBelongAndGet(id, ctx); false == pass {
			return
		}
	} else {
		var err error
		accountUser, err = accountModel.NewDao().SelectUser(id, cf.GetUserId(ctx))
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		account, err = accountUser.GetAccount()
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
	}
	return account, accountUser, true
}

func (cf *contextFunc) GetAccountUserByParam(ctx *gin.Context) (
	accountUser accountModel.User, account accountModel.Account, pass bool,
) {
	id, ok := cf.GetParamId(ctx)
	if false == ok {
		return
	}
	err := accountUser.SelectById(id)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if account, _, pass = CheckFunc.AccountBelongAndGet(accountUser.AccountId, ctx); false == pass {
		return
	}
	return accountUser, account, true
}

func (cf *contextFunc) GetCategoryByParam(ctx *gin.Context) (
	category categoryModel.Category, pass bool,
) {
	id, ok := cf.GetUintParamByKey("id", ctx)
	if false == ok {
		return
	}
	var err error
	category, err = categoryModel.NewDao().SelectById(id)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if pass = CheckFunc.AccountBelong(category.AccountId, ctx); false == pass {
		return
	}
	return category, true
}
