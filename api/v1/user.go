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
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songzhibin97/gkit/egroup"
	"gorm.io/gorm"
	"time"
)

type UserApi struct {
}

type _userPublic interface {
	Login(ctx *gin.Context)
	Register(ctx *gin.Context)
	UpdatePassword(ctx *gin.Context)
}

type _userBase interface {
	responseAndMaskUserInfo(userModel.UserInfo) response.UserInfo
	UpdatePassword(ctx *gin.Context)
	UpdateInfo(ctx *gin.Context)
	SetCurrentAccount(ctx *gin.Context)
	SendCaptchaEmail(ctx *gin.Context)
	Home(ctx *gin.Context)
}

type _userFriend interface {
	GetFriendList(ctx *gin.Context)
	responseUserFriendInvitation(userModel.FriendInvitation) (response.UserFriendInvitation, error)
	CreateFriendInvitation(ctx *gin.Context)
	getFriendInvitationByParam(ctx *gin.Context) (userModel.FriendInvitation, bool)
	AcceptFriendInvitation(ctx *gin.Context)
	RefuseFriendInvitation(ctx *gin.Context)
	GetFriendInvitationList(ctx *gin.Context)
}

type _userConfig interface {
	GetTransactionShareConfig(ctx *gin.Context)
	UpdateTransactionShareConfig(ctx *gin.Context)
}

func (p *PublicApi) Login(ctx *gin.Context) {
	var requestData request.UserLogin
	var err error
	// 处理错误方法
	var loginFailResponseFunc = func() {
		if err != nil {
			key := global.Cache.GetKey(constant.LoginFailCount, requestData.Email)
			count, existCache := global.Cache.Get(key)
			if existCache {
				if intCount, ok := count.(int); ok {
					if intCount > 5 {
						response.FailToError(ctx, errors.New("错误次数过的，请稍后再试"))
						return
					} else {
						_ = global.Cache.Increment(key, 1)
					}
				} else {
					panic("cache计数数据转断言int失败")
				}
			} else {
				global.Cache.Set(key, 1, time.Hour*12)
			}
			response.FailToError(ctx, err)
			return
		}
	}
	defer loginFailResponseFunc()
	// 请求数据获取与校验
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		return
	}
	if false == captchaStore.Verify(requestData.CaptchaId, requestData.Captcha, true) {
		response.FailWithMessage("验证码错误", ctx)
		return
	}

	client := contextFunc.GetClient(ctx)
	// 开始处理
	var user userModel.User
	responseData := response.Login{}
	transactionFunc := func(tx *gorm.DB) error {
		var clientBaseInfo userModel.UserClientBaseInfo
		user, clientBaseInfo, responseData.Token, err = userService.Base.Login(
			requestData.Email, requestData.Password, client, tx,
		)
		if err != nil {
			return err
		}
		err = responseData.User.SetData(user)
		if err != nil {
			return err
		}
		// 当前客户端操作账本
		if clientBaseInfo.CurrentAccountId != 0 {
			var account accountModel.Account
			account, err = accountModel.NewDao().SelectById(clientBaseInfo.CurrentAccountId)
			if err != nil {
				return err
			}
			err = responseData.CurrentAccount.SetDataFromAccount(account)
			if err != nil {
				return err
			}
		}
		// 当前客户端操作共享账本
		if clientBaseInfo.CurrentShareAccountId != 0 {
			var accountUser accountModel.User
			accountUser, err = accountModel.NewDao().SelectUser(clientBaseInfo.CurrentShareAccountId, user.ID)
			if err != nil && false == errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			err = responseData.CurrentShareAccount.SetData(accountUser)
			if err != nil {
				return err
			}
		}
		return err
	}

	if err = global.GvaDb.Transaction(transactionFunc); err != nil {
		err = errors.New("用户名不存在或者密码错误")
		return
	}
	if responseData.Token == "" {
		err = errors.New("token获取失败")
		return
	}
	if err == nil {
		response.OkWithDetailed(responseData, "登录成功", ctx)
	}
}

func (p *PublicApi) Register(ctx *gin.Context) {
	var requestData request.UserRegister
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	err = commonService.CheckEmailCaptcha(requestData.Email, requestData.Captcha)
	if responseError(err, ctx) {
		return
	}

	data := userModel.AddData{Username: requestData.Username, Password: requestData.Password, Email: requestData.Email}
	var user userModel.User
	var token string
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			user, err = userService.Base.Register(data, tx)
			if err != nil {
				return err
			}
			//注册成功 获取token
			customClaims := commonService.MakeCustomClaims(user.ID)
			token, err = commonService.GenerateJWT(customClaims)
			if err != nil {
				return err
			}
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}
	// 发送不成功不影响主流程
	_ = thirdpartyService.SendNotificationEmail(constant.NotificationOfRegistrationSuccess, &user)

	responseData := response.Register{Token: token}
	err = responseData.User.SetData(user)
	if responseError(err, ctx) {
		return
	}
	response.OkWithDetailed(responseData, "注册成功", ctx)
}

func (p *PublicApi) UpdatePassword(ctx *gin.Context) {
	var requestData request.UserForgetPassword
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	err := commonService.CheckEmailCaptcha(requestData.Email, requestData.Captcha)
	if responseError(err, ctx) {
		return
	}
	user, err := userModel.NewDao().SelectByEmail(requestData.Email)
	if responseError(err, ctx) {
		return
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return userService.Base.UpdatePassword(user, requestData.Password, tx)
		},
	)
	// 发送不成功不影响主流程
	_ = thirdpartyService.SendNotificationEmail(constant.NotificationOfUpdatePassword, &user)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (u *UserApi) UpdatePassword(ctx *gin.Context) {
	var requestData request.UserUpdatePassword
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	err = commonService.CheckEmailCaptcha(user.Email, requestData.Captcha)
	if responseError(err, ctx) {
		return
	}

	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return userService.Base.UpdatePassword(user, requestData.Password, tx)
		},
	)
	// 发送不成功不影响主流程
	_ = thirdpartyService.SendNotificationEmail(constant.NotificationOfUpdatePassword, &user)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (u *UserApi) UpdateInfo(ctx *gin.Context) {
	var requestData request.UserUpdateInfo
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
	err = global.GvaDb.Model(&user).Update("username", requestData.Username).Error
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (u *UserApi) SearchUser(ctx *gin.Context) {
	var requestData request.UserSearch
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	condition := userModel.Condition{
		Id:                 requestData.Id,
		LikePrefixUsername: &requestData.Username,
		Offset:             requestData.Offset,
		Limit:              requestData.Limit,
	}
	var list []userModel.UserInfo
	list, err = userModel.NewDao().SelectUserInfoByCondition(condition)
	var responseData response.List[response.UserInfo]
	responseData.List = make([]response.UserInfo, len(list), len(list))
	for i := 0; i < len(responseData.List); i++ {
		responseData.List[i].SetMaskData(list[i])
	}
	response.OkWithData(responseData, ctx)
}

func (u *UserApi) SetCurrentAccount(ctx *gin.Context) {
	var requestData request.Id
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account, accountUser, pass := checkFunc.AccountBelongAndGet(requestData.Id, ctx)
	if false == pass {
		return
	}

	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return userService.Base.SetClientAccount(accountUser, contextFunc.GetClient(ctx), account, tx)
		},
	)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (u *UserApi) SetCurrentShareAccount(ctx *gin.Context) {
	var requestData request.Id
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account, accountUser, pass := checkFunc.AccountBelongAndGet(requestData.Id, ctx)
	if false == pass {
		return
	}
	if account.Type != accountModel.TypeShare {
		response.FailToError(ctx, global.ErrAccountType)
		return
	}

	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return userService.Base.SetClientShareAccount(accountUser, contextFunc.GetClient(ctx), account, tx)
		},
	)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (u *UserApi) SendCaptchaEmail(ctx *gin.Context) {
	var requestData request.UserSendEmail
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	if false == captchaStore.Verify(requestData.CaptchaId, requestData.Captcha, true) {
		response.FailWithMessage("验证码错误", ctx)
		return
	}

	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}

	err = thirdpartyService.SendCaptchaEmail(user.Email, requestData.Type)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.ExpirationTime{ExpirationTime: global.Config.Captcha.EmailCaptchaTimeOut}, ctx)
}

func (u *UserApi) Home(ctx *gin.Context) {
	var requestData request.UserHome
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	pass := checkFunc.AccountBelong(requestData.AccountId, ctx)
	if false == pass {
		return
	}

	group := egroup.WithContext(ctx)
	nowTime := time.Now()
	year, month, day := time.Now().Date()
	var todayData, yesterdayData, weekData, monthData, yearData response.TransactionStatistic
	condition := transactionModel.StatisticCondition{
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{AccountId: requestData.AccountId},
	}
	handelOneTime := func(data *response.TransactionStatistic, startTime time.Time, endTime time.Time) error {
		condition.StartTime = startTime
		condition.EndTime = endTime
		result, err := transactionModel.NewDao().GetIeStatisticByCondition(nil, condition, nil)
		if err != nil {
			return err
		}
		*data = response.TransactionStatistic{
			IncomeExpenseStatistic: result,
			StartTime:              startTime.Unix(),
			EndTime:                endTime.Unix(),
		}
		return nil
	}
	handelGoroutineOne := func() error {
		var err error
		//今日统计
		if err = handelOneTime(
			&todayData,
			time.Date(year, month, day, 0, 0, 0, 0, time.Local),
			time.Date(year, month, day, 23, 59, 59, 0, time.Local),
		); err != nil {
			return err
		}
		//昨日统计
		if err = handelOneTime(
			&yesterdayData,
			time.Date(year, month, day-1, 0, 0, 0, 0, time.Local),
			time.Date(year, month, day-1, 23, 59, 59, 0, time.Local),
		); err != nil {
			return err
		}
		//周统计
		if err = handelOneTime(
			&weekData,
			util.Time.GetFirstSecondOfMonday(nowTime),
			time.Date(year, month, day, 23, 59, 59, 0, time.Local),
		); err != nil {
			return err
		}
		return err
	}

	handelGoroutineTwo := func() error {
		var err error
		//月统计
		if err = handelOneTime(
			&monthData,
			util.Time.GetFirstSecondOfMonth(nowTime),
			time.Date(year, month, day, 23, 59, 59, 0, time.Local),
		); err != nil {
			return err
		}
		//年统计
		if err = handelOneTime(
			&yearData,
			util.Time.GetFirstSecondOfYear(nowTime),
			time.Date(year, month, day, 23, 59, 59, 0, time.Local),
		); err != nil {
			return err
		}
		return err
	}
	group.Go(handelGoroutineOne)
	group.Go(handelGoroutineTwo)
	// 等待所有 Goroutine 完成
	if err := group.Wait(); responseError(err, ctx) {
		return
	}
	// 处理响应
	responseData := &response.UserHome{}
	responseData.HeaderCard = &response.UserHomeHeaderCard{TransactionStatistic: &monthData}
	responseData.TimePeriodStatistics = &response.UserHomeTimePeriodStatistics{
		TodayData: &todayData, YesterdayData: &yesterdayData, WeekData: &weekData, YearData: &yearData,
	}
	response.OkWithData(responseData, ctx)
}

func (u *UserApi) GetTransactionShareConfig(ctx *gin.Context) {
	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	var config userModel.TransactionShareConfig
	if err = config.SelectByUserId(user.ID); responseError(err, ctx) {
		return
	}
	var responseData response.UserTransactionShareConfig
	err = responseData.SetData(config)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (u *UserApi) UpdateTransactionShareConfig(ctx *gin.Context) {
	var err error
	// 处理请求数据
	var requestData request.UserTransactionShareConfigUpdate
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	// 处理
	config, err := user.GetTransactionShareConfig()
	if responseError(err, ctx) {
		return
	}
	flag, err := request.GetFlagByFlagName(requestData.Flag)
	if responseError(err, ctx) {
		return
	}

	if requestData.Status {
		err = config.OpenDisplayFlag(flag, global.GvaDb)
	} else {
		err = config.ClosedDisplayFlag(flag, global.GvaDb)
	}
	// 响应
	if responseError(err, ctx) {
		return
	}
	config, err = user.GetTransactionShareConfig()
	if responseError(err, ctx) {
		return
	}
	var responseData response.UserTransactionShareConfig
	err = responseData.SetData(config)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (u *UserApi) responseAndMaskUserInfo(data userModel.UserInfo) response.UserInfo {
	return response.UserInfo{
		Id:       data.ID,
		Username: data.Username,
		Email:    util.Str.MaskEmail(data.Email),
	}
}

func (u *UserApi) GetFriendList(ctx *gin.Context) {
	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	var friendList []userModel.Friend
	friendList, err = userModel.NewDao().SelectFriendList(user.ID)

	responseData := make([]response.UserInfo, len(friendList), len(friendList))
	var info userModel.UserInfo
	for i := 0; i < len(responseData); i++ {
		info, err = friendList[i].GetFriendInfo()
		if responseError(err, ctx) {
			return
		}
		responseData[i].SetMaskData(info)

	}
	response.OkWithData(response.List[response.UserInfo]{List: responseData}, ctx)
}

func (u *UserApi) responseUserFriendInvitation(data userModel.FriendInvitation) (
	responseData response.UserFriendInvitation, err error,
) {
	var inviterInfo userModel.UserInfo
	var inviteeInfo userModel.UserInfo
	inviterInfo, err = data.GetInviterInfo()
	if err != nil {
		return
	}
	inviteeInfo, err = data.GetInviteeInfo()
	if err != nil {
		return
	}
	responseData = response.UserFriendInvitation{
		Id:         data.ID,
		CreateTime: data.CreatedAt.Unix(),
	}
	responseData.Inviter.SetMaskData(inviterInfo)
	responseData.Inviter.SetMaskData(inviteeInfo)
	return
}

func (u *UserApi) CreateFriendInvitation(ctx *gin.Context) {
	var requestData request.UserCreateFriendInvitation
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	// 处理
	var invitation userModel.FriendInvitation
	txFunc := func(tx *gorm.DB) error {
		var invitee userModel.User
		invitee, err = userModel.NewDao(tx).SelectById(requestData.Invitee)
		if err != nil {
			return err
		}
		invitation, err = userService.Friend.CreateInvitation(user, invitee, tx)
		if err != nil {
			return err
		}
		return nil
	}
	err = global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.UserFriendInvitation
	responseData, err = u.responseUserFriendInvitation(invitation)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (u *UserApi) getFriendInvitationByParam(ctx *gin.Context) (result userModel.FriendInvitation, isSuccess bool) {
	id, pass := contextFunc.GetUintParamByKey("id", ctx)
	if false == pass {
		return
	}
	if pass, result = checkFunc.FriendInvitationBelongAndGet(id, ctx); pass == false {
		return
	}
	isSuccess = true
	return
}

func (u *UserApi) AcceptFriendInvitation(ctx *gin.Context) {
	invitation, pass := u.getFriendInvitationByParam(ctx)
	if false == pass {
		return
	}
	if invitation.Invitee != contextFunc.GetUserId(ctx) {
		response.FailToError(ctx, errors.New("非被邀请者！"))
		return
	}
	txFunc := func(tx *gorm.DB) error {
		_, _, err := userService.Friend.AcceptInvitation(&invitation, tx)
		return err
	}
	err := global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	var responseData response.UserFriendInvitation
	responseData, err = u.responseUserFriendInvitation(invitation)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (u *UserApi) RefuseFriendInvitation(ctx *gin.Context) {
	invitation, pass := u.getFriendInvitationByParam(ctx)
	if false == pass {
		return
	}
	if invitation.Invitee != contextFunc.GetUserId(ctx) {
		response.FailToError(ctx, errors.New("非被邀请者！"))
		return
	}
	txFunc := func(tx *gorm.DB) error {
		err := userService.Friend.RefuseInvitation(&invitation, tx)
		return err
	}
	err := global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	var responseData response.UserFriendInvitation
	responseData, err = u.responseUserFriendInvitation(invitation)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (u *UserApi) GetFriendInvitationList(ctx *gin.Context) {
	var requestData request.UserGetFriendInvitation
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	var list []userModel.FriendInvitation
	if requestData.IsInvite {
		list, err = userModel.NewDao().SelectFriendInvitationList(&user.ID, nil)
	} else {
		list, err = userModel.NewDao().SelectFriendInvitationList(nil, &user.ID)
	}

	responseData := make([]response.UserFriendInvitation, len(list), len(list))
	for i := 0; i < len(responseData); i++ {
		responseData[i], err = u.responseUserFriendInvitation(list[i])
		if responseError(err, ctx) {
			return
		}
	}
	response.OkWithData(response.List[response.UserFriendInvitation]{List: responseData}, ctx)
}

func (u *UserApi) GetAccountInvitationList(ctx *gin.Context) {
	var err error
	var requestData request.UserGetAccountInvitationList
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	// 查询
	condition := accountModel.NewUserInvitationCondition(requestData.Limit, requestData.Offset)
	condition.SetInviteeId(contextFunc.GetUserId(ctx))
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
