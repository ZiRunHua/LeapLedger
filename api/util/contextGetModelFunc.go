package util

import (
	"KeepAccount/api/response"
	accountModel "KeepAccount/model/account"
	transactionModel "KeepAccount/model/transaction"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type _getModelByContextFunc interface {
	// 获取交易从urlParam
	GetTransByParam(ctx *gin.Context) (*transactionModel.Transaction, bool)
}

func (cf *contextFunc) GetTransByParam(ctx *gin.Context) (result transactionModel.Transaction, pass bool) {
	id, ok := cf.GetParamId(ctx)
	if false == ok {
		response.FailToError(ctx, errors.New("error param id"))
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
	id, ok := cf.GetUintParamByKey("id", ctx)
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
		response.FailToError(ctx, errors.New("error param id"))
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
