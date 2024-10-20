package v1

import (
	"context"
	"errors"
	"time"

	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"KeepAccount/util/timeTool"
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/egroup"
	"gorm.io/gorm"
)

type AccountApi struct {
}

// GetOne
//
//	@Tags		Account
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"Account ID"
//	@Success	200	{object}	response.Data{Data=response.AccountDetail}
//	@Router		/account/{id} [get]
func (a *AccountApi) GetOne(ctx *gin.Context) {
	_, accountUser, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	var responseData response.AccountDetail
	err := responseData.SetData(accountUser)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// CreateOne
//
//	@Tags		Account
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int							true	"Account ID"
//	@Param		body	body		request.AccountCreateOne	true	"Account data"
//	@Success	200		{object}	response.Data{Data=response.AccountDetail}
//	@Router		/account [post]
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
	var _ accountModel.Account
	_, aUser, err := accountService.CreateOne(
		user, accountService.NewCreateData(
			requestData.Name,
			requestData.Icon,
			requestData.Type,
			requestData.Location,
		), ctx,
	)
	if responseError(err, ctx) {
		return
	}
	var responseData response.AccountDetail
	err = responseData.SetData(aUser)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// Update
//
//	@Tags		Account
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int							true	"Account ID"
//	@Param		body	body		request.AccountUpdateOne	true	"Account data"
//	@Success	200		{object}	response.Data{Data=response.AccountDetail}
//	@Router		/account/{id} [put]
func (a *AccountApi) Update(ctx *gin.Context) {
	var requestData request.AccountUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account, accountUser, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}

	err := accountService.Update(
		account, accountUser,
		accountModel.AccountUpdateData{Name: requestData.Name, Icon: requestData.Icon, Type: requestData.Type}, ctx,
	)
	if responseError(err, ctx) {
		return
	}
	// response
	account, err = accountModel.NewDao().SelectById(account.ID)
	if responseError(err, ctx) {
		return
	}
	var responseData response.AccountDetail
	err = responseData.SetDataFromAccount(account)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// Delete
//
//	@Tags		Account
//	@Produce	json
//	@Param		id	path		int													true	"Account ID"
//	@Success	200	{object}	response.Data{Data=response.UserCurrentClientInfo}	"new current client info"
//	@Router		/account/{id} [delete]
func (a *AccountApi) Delete(ctx *gin.Context) {
	account, accountUser, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	txFunc := func(tx *gorm.DB) error {
		return accountService.Delete(account, accountUser, context.WithValue(ctx, cus.Db, tx))
	}

	if err := db.Db.Transaction(txFunc); responseError(err, ctx) {
		return
	}
	// response可以已被更新的当前客户端信息
	var responseData response.UserCurrentClientInfo
	clientInfo, err := contextFunc.GetUserCurrentClientInfo(ctx)
	if responseError(err, ctx) {
		return
	}
	err = responseData.SetData(clientInfo)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetList
//
//	@Tags		Account
//	@Produce	json
//	@Success	200	{object}	response.Data{Data=response.List[response.AccountDetail]{}}
//	@Router		/account/list [get]
func (a *AccountApi) GetList(ctx *gin.Context) {
	var accountUserList []accountModel.User
	err := db.Db.Where("user_id = ?", contextFunc.GetUserId(ctx)).Find(&accountUserList).Error
	if responseError(err, ctx) {
		return
	}
	var responseData response.AccountDetailList
	err = responseData.SetData(accountUserList)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.List[response.AccountDetail]{List: responseData}, ctx)
}

// GetListByType
//
//	@Tags		Account
//	@Produce	json
//	@Param		type	path		int	true	"Account type"
//	@Success	200		{object}	response.Data{Data=response.List[response.AccountDetail]{}}
//	@Router		/account/list/{type} [get]
func (a *AccountApi) GetListByType(ctx *gin.Context) {
	t := contextFunc.GetAccountType(ctx)
	list, err := accountModel.NewDao().SelectUserListByUserAndAccountType(contextFunc.GetUserId(ctx), t)

	if responseError(err, ctx) {
		return
	}
	var responseData response.AccountDetailList
	err = responseData.SetData(list)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.List[response.AccountDetail]{List: responseData}, ctx)
}

// CreateOneByTemplate
//
//	@Tags		Account/Template
//	@Produce	json
//	@Param		id	path		int	true	"Template ID"
//	@Success	200	{object}	response.Data{Data=response.AccountDetail}
//	@Router		/account/form/template/{id} [post]
func (a *AccountApi) CreateOneByTemplate(ctx *gin.Context) {
	id, ok := contextFunc.GetUintParamByKey("id", ctx)
	if false == ok {
		return
	}
	tmpAccount, err := accountModel.NewDao().SelectById(id)
	if responseError(err, ctx) {
		return
	}
	var user userModel.User
	user, err = contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	account, err := templateService.CreateAccount(user, tmpAccount, ctx)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountDetail
	err = responseData.SetDataFromAccount(account)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetAccountTemplateList
//
//	@Tags		Account/Template
//	@Produce	json
//	@Success	200	{object}	response.Data{Data=response.AccountTemplateList}
//	@Router		/account/template/list [get]
func (a *AccountApi) GetAccountTemplateList(ctx *gin.Context) {
	list, err := templateService.GetListByRank(ctx)
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
				Type: account.Type,
			},
		)
	}
	response.OkWithData(responseData, ctx)
}

// InitCategoryByTemplate
//
//	@Tags		Account/Template
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int									true	"Account ID"
//	@Param		body		body		request.AccountTransCategoryInit	true	"init data"
//	@Success	200			{object}	response.Data{Data=response.AccountDetail}
//	@Router		/account/{accountId}/transaction/category/init [post]
func (a *AccountApi) InitCategoryByTemplate(ctx *gin.Context) {
	var err error
	var requestData request.AccountTransCategoryInit
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	var template accountModel.Account
	account, _, pass := contextFunc.GetAccountByParam(ctx, true)
	if pass == false {
		return
	}
	if err = db.Db.First(&template, requestData.TemplateId).Error; responseError(err, ctx) {
		return
	}

	err = templateService.CreateCategory(account, template, ctx)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountDetail
	err = responseData.SetDataFromAccount(account)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetUserInvitationList
//
//	@Tags		Account/User/Invitation
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.AccountGetUserInvitationList	true	"query param"
//	@Success	200		{object}	response.Data{Data=response.List[response.AccountUserInvitation]{}}
//	@Router		/account/user/invitation/list [get]
func (a *AccountApi) GetUserInvitationList(ctx *gin.Context) {
	var err error
	var requestData request.AccountGetUserInvitationList
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if pass := checkFunc.AccountBelong(requestData.AccountId, ctx); false == pass {
		return
	}
	// 设置查询条件
	condition := accountModel.NewUserInvitationCondition(requestData.Limit, requestData.Offset)
	condition.SetAccountId(requestData.AccountId)
	if requestData.Role != nil {
		condition.SetPermission(requestData.Role.ToUserPermission())
	}
	if requestData.Invitee != nil {
		condition.SetInviteeId(*requestData.Invitee)
	}

	var list []accountModel.UserInvitation
	list, err = accountModel.NewDao().SelectUserInvitationByCondition(*condition)
	if responseError(err, ctx) {
		return
	}
	// response
	responseData := make([]response.AccountUserInvitation, len(list), len(list))
	for i := 0; i < len(responseData); i++ {
		err = responseData[i].SetData(list[i])
		if responseError(err, ctx) {
			return
		}
	}
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.List[response.AccountUserInvitation]{List: responseData}, ctx)
}

// CreateAccountUserInvitation
//
//	@Tags		Account/User/Invitation
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int										true	"Account ID"
//	@Param		body		body		request.AccountGetUserInvitationList	true	"invitation data"
//	@Success	200			{object}	response.Data{Data=response.AccountUserInvitation}
//	@Router		/account/{accountId}/user/invitation [post]
func (a *AccountApi) CreateAccountUserInvitation(ctx *gin.Context) {
	var err error
	var requestData request.AccountCreateOneUserInvitation
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	// 获取数据
	account, inviter, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	var invitee userModel.User
	invitee, err = userModel.NewDao().SelectById(requestData.Invitee)
	if responseError(err, ctx) {
		return
	}
	// handle
	var invitation accountModel.UserInvitation
	// 不为创建者不可设置角色 只能取默认值
	if requestData.Role == nil || inviter.GetRole() != accountModel.Creator {
		invitation, err = accountService.Share.CreateUserInvitation(
			account, inviter, invitee, nil, ctx,
		)
	} else {
		rolePermission := requestData.Role.ToUserPermission()
		invitation, err = accountService.Share.CreateUserInvitation(
			account, inviter, invitee, &rolePermission, ctx,
		)
	}
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountUserInvitation
	err = responseData.SetData(invitation)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) getAcceptUserInvitationByParam(ctx *gin.Context) (accountModel.UserInvitation, bool) {
	var invitation accountModel.UserInvitation
	id, isSuccess := contextFunc.GetUintParamByKey("id", ctx)
	if isSuccess == false {
		return invitation, false
	}
	err := db.Db.First(&invitation, id).Error
	if responseError(err, ctx) {
		return invitation, false
	}
	return invitation, true
}

// AcceptAccountUserInvitation
//
//	@Tags		Account/User/Invitation
//	@Produce	json
//	@Param		id	path		int	true	"Invitation ID"
//	@Success	200	{object}	response.Data{Data=response.AccountUserInvitation}
//	@Router		/account/user/invitation/{id}/accept [put]
func (a *AccountApi) AcceptAccountUserInvitation(ctx *gin.Context) {
	invitation, pass := a.getAcceptUserInvitationByParam(ctx)
	if false == pass {
		return
	}
	if contextFunc.GetUserId(ctx) != invitation.Invitee {
		response.Forbidden(ctx)
		return
	}
	// handle
	txFunc := func(tx *gorm.DB) error {
		_, err := invitation.Accept(tx)
		return err
	}
	err := db.Db.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountUserInvitation
	err = responseData.SetData(invitation)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// RefuseAccountUserInvitation
//
//	@Tags		Account/User/Invitation
//	@Produce	json
//	@Param		id	path		int	true	"Invitation ID"
//	@Success	200	{object}	response.Data{Data=response.AccountUserInvitation}
//	@Router		/account/user/invitation/{id}/refuse [put]
func (a *AccountApi) RefuseAccountUserInvitation(ctx *gin.Context) {
	invitation, pass := a.getAcceptUserInvitationByParam(ctx)
	if false == pass {
		return
	}
	if contextFunc.GetUserId(ctx) != invitation.Invitee {
		response.Forbidden(ctx)
		return
	}
	// handle
	txFunc := func(tx *gorm.DB) error {
		err := invitation.Refuse(tx)
		return err
	}
	err := db.Db.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountUserInvitation
	err = responseData.SetData(invitation)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// UpdateUser
//
//	@Tags		Account/User
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int							true	"Account ID"
//	@Param		id			path		int							true	"Account User ID"
//	@Param		body		body		request.AccountUpdateUser	true	"account user data"
//	@Success	200			{object}	response.Data{Data=response.AccountUser}
//	@Router		/account/{accountId}/user/{id} [put]
func (a *AccountApi) UpdateUser(ctx *gin.Context) {
	var requestData request.AccountUpdateUser
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	accountUser, _, pass := contextFunc.GetAccountUserByParam(ctx)
	if false == pass {
		return
	}
	if contextFunc.GetAccountId(ctx) != accountUser.AccountId {
		response.FailToParameter(ctx, global.ErrAccountId)
		return
	}
	operator, err := accountModel.NewDao().SelectUser(accountUser.AccountId, contextFunc.GetUserId(ctx))
	if responseError(err, ctx) {
		return
	}
	var result accountModel.User
	result, err = accountService.UpdateUser(accountUser, operator, requestData.GetUpdateData(), ctx)
	if responseError(err, ctx) {
		return
	}
	var responseData response.AccountUser
	err = responseData.SetData(result)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetUserList
//
//	@Tags		Account/User
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Success	200			{object}	response.Data{Data=response.List[response.AccountUser]{}}
//	@Router		/account/{accountId}/user/list [get]
func (a *AccountApi) GetUserList(ctx *gin.Context) {
	list, err := accountModel.NewDao().SelectUserListByAccountId(contextFunc.GetAccountId(ctx))
	// response
	responseData := make([]response.AccountUser, len(list), len(list))
	for i := 0; i < len(responseData); i++ {
		err = responseData[i].SetData(list[i])
		if responseError(err, ctx) {
			return
		}
	}
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.List[response.AccountUser]{List: responseData}, ctx)
}

// GetUserInfo
//
//	@Tags		Account/User
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int							true	"Account User ID"
//	@Param		body	body		request.AccountGetUserInfo	true	"query param"
//	@Success	200		{object}	response.Data{Data=response.AccountUserInfo}
//	@Router		/account/{accountId}/user/{id}/info [get]
func (a *AccountApi) GetUserInfo(ctx *gin.Context) {
	var requestData request.AccountGetUserInfo
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	accountUser, account, nowTime := contextFunc.GetAccountUser(ctx), contextFunc.GetAccount(ctx), contextFunc.GetNowTime(ctx)
	group := egroup.WithContext(ctx)
	var todayTotal, monthTotal *global.IEStatisticWithTime
	var recentTrans *response.TransactionDetailList
	for _, infoType := range requestData.Types {
		switch infoType {
		case request.TodayTransTotal:
			result, err := a.getTransTotal(account, &[]uint{accountUser.UserId}, nowTime, nowTime)
			todayTotal = &result
			if responseError(err, ctx) {
				return
			}

		case request.CurrentMonthTransTotal:
			result, err := a.getTransTotal(
				account, &[]uint{accountUser.UserId}, timeTool.GetFirstSecondOfMonth(nowTime), nowTime,
			)
			monthTotal = &result
			if responseError(err, ctx) {
				return
			}
		case request.RecentTrans:
			result, err := a.getTrans(account, &[]uint{accountUser.UserId}, 5, 0)
			if responseError(err, ctx) {
				return
			}
			recentTrans = &response.TransactionDetailList{}
			err = recentTrans.SetData(result)
			if responseError(err, ctx) {
				return
			}
		}
	}
	if err := group.Wait(); responseError(err, ctx) {
		return
	}
	responseData := &response.AccountUserInfo{
		TodayTransTotal: todayTotal, CurrentMonthTransTotal: monthTotal, RecentTrans: recentTrans,
	}

	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) getTransTotal(
	account accountModel.Account, UserIds *[]uint, start time.Time, end time.Time,
) (result global.IEStatisticWithTime, err error) {
	condition := transactionModel.StatisticCondition{
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{
			AccountId: account.ID,
			UserIds:   UserIds,
		},
		StartTime: start,
		EndTime:   end,
	}
	dao := transactionModel.NewStatisticDao()

	result.Income, err = dao.GetTotalByCondition(constant.Income, condition)
	if err != nil {
		return
	}
	result.Expense, err = dao.GetTotalByCondition(constant.Expense, condition)
	if err != nil {
		return
	}
	result.StartTime = condition.StartTime
	result.EndTime = condition.EndTime
	return
}

func (a *AccountApi) getTrans(
	account accountModel.Account, UserIds *[]uint, limit, offset int,
) ([]transactionModel.Transaction, error) {
	condition := transactionModel.Condition{
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{
			AccountId: account.ID,
			UserIds:   UserIds,
		},
	}
	return transactionModel.NewDao().GetListByCondition(condition, offset, limit)
}

// GetUserConfig
//
//	@Tags		Account/User/Config
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Success	200			{object}	response.Data{Data=response.AccountUserConfig}
//	@Router		/account/{accountId}/user/config [get]
func (a *AccountApi) GetUserConfig(ctx *gin.Context) {
	_, accountUser, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	config, err := accountUser.GetConfig()
	if responseError(err, ctx) {
		return
	}
	var responseData response.AccountUserConfig
	err = responseData.SetData(config)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

var AccountConfigFlagMap = map[string]interface{}{"SyncMappingAccount": accountModel.Flag_Trans_Sync_Mapping_Account}

func (a *AccountApi) getUserConfigFlagByCtx(ctx *gin.Context) interface{} {
	return AccountConfigFlagMap[ctx.Param("type")]
}

// UpdateUserConfigFlag
//
//	@Tags		Account/User/Config
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int									true	"Account ID"
//	@Param		body		body		request.AccountUserConfigFlagUpdate	true	"config data"
//	@Success	200			{object}	response.Data{Data=response.AccountUserConfig}
//	@Router		/account/{accountId}/user/config/{flag} [put]
func (a *AccountApi) UpdateUserConfigFlag(ctx *gin.Context) {
	var requestData request.AccountUserConfigFlagUpdate
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	_, accountUser, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	// handle
	userConfig, err := accountUser.GetConfig()
	if responseError(err, ctx) {
		return
	}
	err = db.Db.Transaction(
		func(tx *gorm.DB) error {
			err = userConfig.ForShare(tx)
			if err != nil {
				return err
			}
			if requestData.Status {
				err = userConfig.OpenUserConfigFlag(a.getUserConfigFlagByCtx(ctx), tx)
			} else {
				err = userConfig.CloseUserConfigFlag(a.getUserConfigFlagByCtx(ctx), tx)
			}
			return nil
		},
	)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountUserConfig
	err = responseData.SetData(userConfig)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetAccountMappingList
//
//	@Tags		Account/Mapping
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Success	200			{object}	response.Data{Data=response.List[response.AccountMapping]}
//	@Router		/account/{accountId}/mapping/list [get]
func (a *AccountApi) GetAccountMappingList(ctx *gin.Context) {
	account, _, pass := contextFunc.GetAccountByParam(ctx, true)
	if pass == false {
		return
	}
	mappingList, err := accountModel.NewDao().SelectMultipleMapping(*accountModel.NewMappingCondition().WithMainId(account.ID))
	if responseError(err, ctx) {
		return
	}
	responseData := make([]response.AccountMapping, len(mappingList), len(mappingList))
	for i := 0; i < len(responseData); i++ {
		err = responseData[i].SetData(mappingList[i])
		if responseError(err, ctx) {
			return
		}
	}
	response.OkWithData(response.List[response.AccountMapping]{List: responseData}, ctx)
}

// GetAccountMapping
//
//	@Tags		Account/Mapping
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Success	200			{object}	response.Data{Data=response.AccountMapping}
//	@Router		/account/{accountId}/mapping [get]
func (a *AccountApi) GetAccountMapping(ctx *gin.Context) {
	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	account, _, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	var mapping accountModel.Mapping
	mapping, err = accountModel.NewDao().SelectMappingByMainAccountAndRelatedUser(account.ID, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Ok(ctx)
		}
		response.FailToError(ctx, err)
		return
	}
	var responseData response.AccountMapping
	err = responseData.SetData(mapping)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// CreateAccountMapping
//
//	@Tags		Account/Mapping
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int						true	"Account ID"
//	@Param		body		body		request.AccountMapping	true	"mapping data"
//	@Success	200			{object}	response.Data{Data=response.AccountMapping}
//	@Router		/account/{accountId}/mapping [post]
func (a *AccountApi) CreateAccountMapping(ctx *gin.Context) {
	var err error
	var requestData request.AccountMapping
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	// data
	var user userModel.User
	if user, err = contextFunc.GetUser(ctx); responseError(err, ctx) {
		return
	}
	var mainAccount, mappingAccount accountModel.Account
	var pass bool
	if mainAccount, _, pass = contextFunc.GetAccountByParam(ctx, true); false == pass {
		return
	}
	if mappingAccount, err = accountModel.NewDao().SelectById(requestData.AccountId); responseError(
		err, ctx,
	) {
		return
	}
	// handle
	mapping, err := accountService.Share.MappingAccount(user, mainAccount, mappingAccount, ctx)
	if responseError(err, ctx) {
		return
	}
	err = categoryService.Task.MappingCategoryToAccountMapping(mapping)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountMapping
	err = responseData.SetData(mapping)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// DeleteAccountMapping
//
//	@Tags		Account/Mapping
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Param		id			path		int	true	"Mapping ID"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/mapping/{id} [delete]
func (a *AccountApi) DeleteAccountMapping(ctx *gin.Context) {
	var err error
	id, pass := contextFunc.GetUintParamByKey("id", ctx)
	if false == pass {
		return
	}
	var user userModel.User
	user, err = contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	// handle
	var mapping accountModel.Mapping
	mapping, err = accountModel.NewDao().SelectMappingById(id)
	if responseError(err, ctx) {
		return
	}
	err = accountService.Share.DeleteAccountMapping(user, mapping, ctx)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountMapping
	err = responseData.SetData(mapping)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// UpdateAccountMapping
//
//	@Tags		Account/Mapping
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int								true	"Account ID"
//	@Param		id			path		int								true	"Mapping ID"
//	@Param		body		body		request.UpdateAccountMapping	true	"mapping data"
//	@Success	200			{object}	response.Data{Data=response.AccountMapping}
//	@Router		/account/{accountId}/mapping/{id} [put]
func (a *AccountApi) UpdateAccountMapping(ctx *gin.Context) {
	id, pass := contextFunc.GetUintParamByKey("id", ctx)
	if false == pass {
		return
	}
	var requestData request.UpdateAccountMapping
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	var user userModel.User
	user, err = contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	// handle
	var mapping accountModel.Mapping
	var relatedAccount accountModel.Account
	err = db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			dao := accountModel.NewDao(ctx.GetDb())
			mapping, err = dao.SelectMappingById(id)
			if err != nil {
				return err
			}
			relatedAccount, err = dao.SelectById(requestData.RelatedAccountId)
			if err != nil {
				return err
			}
			mapping, err = accountService.Share.UpdateAccountMapping(user, mapping, relatedAccount, ctx)
			if err != nil {
				return err
			}
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.AccountMapping
	err = responseData.SetData(mapping)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetInfo
//
//	@Tags		Account
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Param		type		path		string	true	"Account type"
//	@Success	200			{object}	response.Data{Data=response.AccountInfo}
//	@Router		/account/{accountId}/info/:type [get]
//	@Router		/account/{accountId}/info [get]
func (a *AccountApi) GetInfo(ctx *gin.Context) {
	var types []request.InfoType
	infoType := contextFunc.GetInfoTypeFormParam(ctx)
	if infoType == "" {
		var requestData request.AccountInfo
		if err := ctx.ShouldBindJSON(&requestData); err != nil {
			response.FailToParameter(ctx, err)
			return
		}
		if requestData.Types == nil {
			response.FailToParameter(ctx, errors.New("type"))
			return
		}
		types = *requestData.Types
	} else {
		types = []request.InfoType{infoType}
	}
	account, nowTime := contextFunc.GetAccount(ctx), contextFunc.GetNowTime(ctx)

	var todayTotal, monthTotal *global.IEStatisticWithTime
	var recentTrans *response.TransactionDetailList
	typeHandleFunc := func(infoType request.InfoType, account accountModel.Account) error {
		switch infoType {
		case request.TodayTransTotal:
			result, err := a.getTransTotal(account, nil, nowTime, nowTime.Add(time.Hour*24-time.Second))
			todayTotal = &result
			return err
		case request.CurrentMonthTransTotal:
			result, err := a.getTransTotal(account, nil, timeTool.GetFirstSecondOfMonth(nowTime), nowTime)
			monthTotal = &result
			return err
		case request.RecentTrans:
			result, err := a.getTrans(account, nil, 10, 0)
			if err != nil {
				return err
			}
			recentTrans = &response.TransactionDetailList{}
			err = recentTrans.SetData(result)
			if err != nil {
				return err
			}
			return err
		}
		return nil
	}
	// process and response
	var group *egroup.Group
	if len(types) > 1 {
		group = egroup.WithContext(ctx)
	}
	for i := range types {
		t := types[i]
		if group != nil {
			group.Go(func() error { return typeHandleFunc(t, account) })
		} else {
			err := typeHandleFunc(types[i], account)
			if responseError(err, ctx) {
				return
			}
		}
	}
	if group != nil {
		if err := group.Wait(); responseError(err, ctx) {
			return
		}
	}

	responseData := response.AccountInfo{
		TodayTransTotal:        todayTotal,
		CurrentMonthTransTotal: monthTotal,
		RecentTrans:            recentTrans,
	}
	response.OkWithData(responseData, ctx)
}
