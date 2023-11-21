package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AccountApi struct {
}

func (a *AccountApi) CreateOne(ctx *gin.Context) {
	var requestData request.AccountCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	user, err := contextFunc.GetUser(ctx)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	var account *accountModel.Account
	txFunc := func(tx *gorm.DB) error {
		account, err = accountService.Base.CreateOne(user, requestData.Name, requestData.Icon, tx)
		return err
	}
	if err = global.GvaDb.Transaction(txFunc); responseError(err, ctx) {
		return
	}
	response.OkWithData(response.AccountModelToResponse(account), ctx)
}

func (a *AccountApi) Update(ctx *gin.Context) {
	var requestData request.AccountUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	pass, account := checkFunc.AccountBelong(ctx.Param("id"), ctx)
	if false == pass {
		return
	}
	txFunc := func(tx *gorm.DB) error {
		return accountModel.Dao.NewAccount(tx).Update(
			account, &accountModel.AccountUpdateData{Name: requestData.Name, Icon: requestData.Icon},
		)
	}
	if err := global.GvaDb.Transaction(txFunc); err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (a *AccountApi) Delete(ctx *gin.Context) {
	pass, account := checkFunc.AccountBelong(ctx.Param("id"), ctx)
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
	rows, err := global.GvaDb.Model(accountModel.Account{}).Where("user_id = ?", contextFunc.GetUserId(ctx)).Rows()
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
			responseData.List, *response.AccountModelToResponse(&account),
		)
	}
	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) GetOne(ctx *gin.Context) {
	pass, account := checkFunc.AccountBelong(ctx.Param("id"), ctx)
	if false == pass {
		return
	}
	response.OkWithData(
		response.AccountModelToResponse(account), ctx,
	)
}

func (a *AccountApi) CreateOneByTemplate(ctx *gin.Context) {
	id, ok := contextFunc.GetUintParamByKey("id", ctx)
	if false == ok {
		return
	}
	user, account := &userModel.User{}, &accountModel.Account{}
	err := global.GvaDb.First(&account, id).Error
	if responseError(err, ctx) {
		return
	}
	if user, err = contextFunc.GetUser(ctx); responseError(err, ctx) {
		return
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			account, err = templateService.CreateAccount(user, account, tx)
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.AccountModelToResponse(account), ctx)
}

func (a *AccountApi) GetAccountTemplateList(ctx *gin.Context) {
	list, err := templateService.GetList()
	if responseError(err, ctx) {
		return
	}
	responseData := response.AccountTemplateList{List: []response.AccountTemplateOne{}}
	for _, account := range list {
		responseData.List = append(
			responseData.List, response.AccountTemplateOne{
				Id:   account.ID,
				Icon: account.Icon,
				Name: account.Name,
			},
		)
	}
	response.OkWithData(responseData, ctx)
}
