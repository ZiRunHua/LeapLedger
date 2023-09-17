package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	accountModel "KeepAccount/model/account"
	"KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	"github.com/gin-gonic/gin"
)

type UserApi struct {
}

func (u *UserApi) SetCurrentAccount(ctx *gin.Context) {
	var requestData request.Id
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account, err := query.FirstByPrimaryKey[*accountModel.Account](requestData.Id)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	var user *userModel.User
	if user, err = contextFunc.GetUser(ctx); err != nil {
		response.FailToError(ctx, err)
		return
	}
	account.BeginTransaction()
	defer account.DeferCommit(ctx)
	if err = userService.SetClientAccount(user, contextFunc.GetClient(ctx), account); err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}
