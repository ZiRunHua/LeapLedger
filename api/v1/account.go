package v1

import (
	"KeepAccount/api/v1/request"
	"KeepAccount/api/v1/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	"github.com/gin-gonic/gin"
)

type AccountApi struct {
}

func (a *AccountApi) CreateOne(ctx *gin.Context) {
	var requestData request.AccountCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	user, err := current.GetUser(ctx)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	account, err := accountService.CreateOne(user, requestData.Name)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.OkWithData(
		response.CreateResponse{
			Id:        account.ID,
			UpdatedAt: account.UpdatedAt.Unix(),
			CreatedAt: account.CreatedAt.Unix(),
		}, ctx,
	)
}

func (a *AccountApi) Update(ctx *gin.Context) {
	var requestData request.Name
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	pass, account := checkAccountBelong(ctx.Param("id"), ctx)
	if false == pass {
		return
	}
	err := global.GvaDb.Model(account).Update("name", requestData.Name).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (a *AccountApi) Delete(ctx *gin.Context) {
	pass, account := checkAccountBelong(ctx.Param("id"), ctx)
	if false == pass {
		return
	}
	err := global.GvaDb.Delete(account).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (a *AccountApi) GetList(ctx *gin.Context) {
	var account accountModel.Account
	rows, err := global.GvaDb.Model(accountModel.Account{}).Where("user_id = ?", current.GetUserId(ctx)).Rows()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	var responseData response.AccountGetAll
	responseData.List = []response.AccountOne{}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &account)
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		responseData.List = append(
			responseData.List, response.AccountOne{
				Id: account.ID, Name: account.Name, CreatedAt: account.CreatedAt.Unix(),
				UpdatedAt: account.UpdatedAt.Unix(),
			},
		)
	}
	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) GetOne(ctx *gin.Context) {
	pass, account := checkAccountBelong(ctx.Param("id"), ctx)
	if false == pass {
		return
	}
	response.OkWithData(
		response.AccountModelToResponse(account), ctx,
	)
}
