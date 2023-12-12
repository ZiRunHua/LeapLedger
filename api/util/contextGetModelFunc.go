package util

import (
	"KeepAccount/api/response"
	accountModel "KeepAccount/model/account"
	transactionModel "KeepAccount/model/transaction"
	"github.com/gin-gonic/gin"
)

type _getModelByContextFunc interface {
	GetTransByParam(ctx *gin.Context) (*transactionModel.Transaction, bool)
}

func (cf *contextFunc) GetTransByParam(ctx *gin.Context) (*transactionModel.Transaction, bool) {
	id, ok := cf.GetParamId(ctx)
	if false == ok {
		return nil, false
	}
	trans := &transactionModel.Transaction{}
	if err := trans.SelectById(id, false); err != nil {
		response.FailToError(ctx, err)
		return nil, false
	}
	if pass, _ := CheckFunc.AccountBelong(trans.AccountId, ctx); false == pass {
		return nil, false
	}
	return trans, true
}

func (cf *contextFunc) GetAndCheckAccountByParam(ctx *gin.Context) (*accountModel.Account, bool) {
	id, ok := cf.GetUintParamByKey("AccountId", ctx)
	if false == ok {
		return nil, false
	}
	pass, account := CheckFunc.AccountBelong(id, ctx)
	if false == pass {
		return nil, false
	}
	return account, true
}
