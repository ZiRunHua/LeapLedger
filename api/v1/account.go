package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/egroup"
	"gorm.io/gorm"
	"time"
)

type AccountApi struct {
}

// GetAccountByParam 获得account通过ctx的Param
func (a *AccountApi) GetAccountByParam(ctx *gin.Context, checkBelong bool) (accountModel.Account, bool) {
	id, pass := contextFunc.GetUintParamByKey("id", ctx)
	if false == pass {
		return accountModel.Account{}, pass
	}
	var account accountModel.Account
	if checkBelong {
		if account, _, pass = checkFunc.AccountBelongAndGet(id, ctx); pass == false {
			return accountModel.Account{}, pass
		}
	} else {
		err := global.GvaDb.First(&account, id).Error
		if responseError(err, ctx) {
			return account, false
		}
	}
	return account, true
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
	var _ accountModel.Account
	var aUser accountModel.User
	txFunc := func(tx *gorm.DB) error {
		_, aUser, err = accountService.Base.CreateOne(user, requestData.Name, requestData.Icon, requestData.Type, tx)
		return err
	}
	if err = global.GvaDb.Transaction(txFunc); responseError(err, ctx) {
		return
	}
	var responseData response.AccountDetail
	err = responseData.SetData(aUser)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

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

	txFunc := func(tx *gorm.DB) error {
		return accountService.Base.Update(
			account, accountUser,
			accountModel.AccountUpdateData{Name: requestData.Name, Icon: requestData.Icon, Type: requestData.Type}, tx,
		)
	}
	err := global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	//响应
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

func (a *AccountApi) Delete(ctx *gin.Context) {
	account, accountUser, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	txFunc := func(tx *gorm.DB) error {
		return accountService.Base.Delete(account, accountUser, tx)
	}

	if err := global.GvaDb.Transaction(txFunc); responseError(err, ctx) {
		return
	}
	// 响应可以已被更新的当前客户端信息
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

func (a *AccountApi) GetList(ctx *gin.Context) {
	var accountUserList []accountModel.User
	err := global.GvaDb.Where("user_id = ?", contextFunc.GetUserId(ctx)).Find(&accountUserList).Error
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

func (a *AccountApi) GetOne(ctx *gin.Context) {
	account, _, pass := contextFunc.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	responseData := response.AccountOne{}
	err := responseData.SetData(account)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(
		response.AccountModelToResponse(&account), ctx,
	)
}

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
	var account accountModel.Account
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			account, err = templateService.CreateAccount(user, tmpAccount, tx)
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.AccountDetail
	err = responseData.SetDataFromAccount(account)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
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
				Type: account.Type,
			},
		)
	}
	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) InitTransCategoryByTemplate(ctx *gin.Context) {
	var err error
	var requestData request.AccountTransCategoryInit
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	account, template := accountModel.Account{}, accountModel.Account{}
	id, pass := contextFunc.GetUintParamByKey("id", ctx)
	if false == pass {
		return
	}
	if account, _, pass = checkFunc.AccountBelongAndGet(id, ctx); pass == false {
		return
	}
	if err = global.GvaDb.First(&template, requestData.TemplateId).Error; responseError(err, ctx) {
		return
	}

	txFunc := func(tx *gorm.DB) error {
		err = templateService.CreateCategory(account, template, tx)
		return err
	}
	err = global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.AccountDetail
	err = responseData.SetDataFromAccount(account)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

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
	//设置查询条件
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
	// 响应
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
	// 处理
	var invitation accountModel.UserInvitation
	txFunc := func(tx *gorm.DB) error {
		// 不为创建者不可设置角色 只能取默认值
		if requestData.Role == nil || inviter.GetRole() != accountModel.Creator {
			invitation, err = accountService.Share.CreateUserInvitation(
				account, inviter, invitee, nil, tx,
			)
			return err
		}
		rolePermission := requestData.Role.ToUserPermission()
		invitation, err = accountService.Share.CreateUserInvitation(
			account, inviter, invitee, &rolePermission, tx,
		)
		return err
	}
	err = global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// 响应
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
	err := global.GvaDb.First(&invitation, id).Error
	if responseError(err, ctx) {
		return invitation, false
	}
	return invitation, true
}

func (a *AccountApi) AcceptAccountUserInvitation(ctx *gin.Context) {
	invitation, pass := a.getAcceptUserInvitationByParam(ctx)
	if false == pass {
		return
	}
	if contextFunc.GetUserId(ctx) != invitation.Invitee {
		response.Forbidden(ctx)
		return
	}
	// 处理
	txFunc := func(tx *gorm.DB) error {
		_, err := invitation.Accept(tx)
		return err
	}
	err := global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.AccountUserInvitation
	err = responseData.SetData(invitation)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) RefuseAccountUserInvitation(ctx *gin.Context) {
	invitation, pass := a.getAcceptUserInvitationByParam(ctx)
	if false == pass {
		return
	}
	if contextFunc.GetUserId(ctx) != invitation.Invitee {
		response.Forbidden(ctx)
		return
	}
	// 处理
	txFunc := func(tx *gorm.DB) error {
		err := invitation.Refuse(tx)
		return err
	}
	err := global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.AccountUserInvitation
	err = responseData.SetData(invitation)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) UpdateUser(ctx *gin.Context) {
	var requestData request.AccountUpdateUser
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	accountUser, account, pass := contextFunc.GetAccountUserByParam(ctx)
	if false == pass {
		return
	}
	operator, err := accountModel.NewDao().SelectUser(account.ID, contextFunc.GetUserId(ctx))
	if responseError(err, ctx) {
		return
	}

	var result accountModel.User
	txFunc := func(tx *gorm.DB) error {
		result, err = accountService.Base.UpdateUser(accountUser, operator, requestData.GetUpdateData(), tx)
		return err
	}
	err = global.GvaDb.Transaction(txFunc)
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

func (a *AccountApi) GetUserList(ctx *gin.Context) {
	account, pass := a.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}
	list, err := accountModel.NewDao().SelectUserListByAccountId(account.ID)
	// 响应
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

func (a *AccountApi) GetUserInfo(ctx *gin.Context) {
	var err error
	var requestData request.AccountGetUserInfo
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	accountUser, account, pass := contextFunc.GetAccountUserByParam(ctx)
	if false == pass {
		return
	}
	group := egroup.WithContext(ctx)
	var todayTotal, monthTotal *global.IncomeExpenseStatistic
	var recentTrans *response.TransactionDetailList
	for _, infoType := range requestData.Types {
		switch infoType {
		case request.TodayTransTotal:
			group.Go(
				func() error {
					result, err := a.getTransTotal(account, &[]uint{accountUser.UserId}, time.Now(), time.Now())
					todayTotal = &result
					return err
				},
			)

		case request.CurrentMonthTransTotal:
			group.Go(
				func() error {
					result, err := a.getTransTotal(
						account, &[]uint{accountUser.UserId}, util.Time.GetFirstSecondOfMonth(time.Now()), time.Now(),
					)
					monthTotal = &result
					return err
				},
			)
		case request.RecentTrans:
			group.Go(
				func() error {
					result, err := a.getTrans(account, &[]uint{accountUser.UserId}, 10, 0)
					if err != nil {
						return err
					}
					recentTrans = &response.TransactionDetailList{}
					err = recentTrans.SetData(result)
					if err != nil {
						return err
					}
					return err
				},
			)
		}
	}
	if err = group.Wait(); responseError(err, ctx) {
		return
	}
	responseData := &response.AccountUserInfo{
		TodayTransTotal: todayTotal, CurrentMonthTransTotal: monthTotal, RecentTrans: recentTrans,
	}

	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) getTransTotal(
	account accountModel.Account, UserIds *[]uint, start time.Time, end time.Time,
) (result global.IncomeExpenseStatistic, err error) {
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
	return transactionModel.NewDao().GetListByCondition(condition, limit, offset)
}

func (a *AccountApi) GetAccountMappingList(ctx *gin.Context) {
	account, pass := a.GetAccountByParam(ctx, true)
	if pass == false {
		return
	}
	mappingList, err := accountModel.NewDao().SelectAllMappingByAccount(account)
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

func (a *AccountApi) CreateAccountMapping(ctx *gin.Context) {
	var err error
	var requestData request.AccountMapping
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	// 获取数据
	var user userModel.User
	if user, err = contextFunc.GetUser(ctx); responseError(err, ctx) {
		return
	}
	var mainAccount, mappingAccount accountModel.Account
	var pass bool
	if mainAccount, pass = a.GetAccountByParam(ctx, true); false == pass {
		return
	}
	if mappingAccount, err = accountModel.NewDao().SelectById(requestData.AccountId); responseError(
		err, ctx,
	) {
		return
	}
	// 处理
	var mapping accountModel.Mapping
	txFunc := func(tx *gorm.DB) error {
		mapping, err = accountService.Share.MappingAccount(user, mainAccount, mappingAccount, tx)
		return err
	}
	err = global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.AccountMapping
	err = responseData.SetData(mapping)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

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
	// 处理
	var mapping accountModel.Mapping
	txFunc := func(tx *gorm.DB) error {
		mapping, err = accountModel.NewDao(tx).SelectMappingById(id)
		if err != nil {
			return err
		}
		err = accountService.Share.DeleteAccountMapping(user, mapping, tx)
		if err != nil {
			return err
		}
		return err
	}
	err = global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.AccountMapping
	err = responseData.SetData(mapping)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

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
	// 处理
	var mapping accountModel.Mapping
	var relatedAccount accountModel.Account
	txFunc := func(tx *gorm.DB) error {
		dao := accountModel.NewDao(tx)
		mapping, err = dao.SelectMappingById(id)
		if err != nil {
			return err
		}
		relatedAccount, err = dao.SelectById(requestData.RelatedAccountId)
		if err != nil {
			return err
		}
		err = accountService.Share.UpdateAccountMapping(user, mapping, relatedAccount, tx)
		if err != nil {
			return err
		}
		return err
	}
	err = global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.AccountMapping
	err = responseData.SetData(mapping)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (a *AccountApi) GetInfo(ctx *gin.Context) {
	//获取信息类型
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
	//查询
	var todayTotal, monthTotal *global.IncomeExpenseStatistic
	var recentTrans *response.TransactionDetailList
	typeHandleFunc := func(infoType request.InfoType, account accountModel.Account) error {
		switch infoType {
		case request.TodayTransTotal:
			result, err := a.getTransTotal(account, nil, time.Now(), time.Now())
			todayTotal = &result
			return err
		case request.CurrentMonthTransTotal:
			result, err := a.getTransTotal(account, nil, util.Time.GetFirstSecondOfMonth(time.Now()), time.Now())
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
	// 处理响应
	account, pass := a.GetAccountByParam(ctx, true)
	if false == pass {
		return
	}

	var group *egroup.Group
	if len(types) > 1 {
		// 启用协程
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
