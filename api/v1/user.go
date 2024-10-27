package v1

import (
	"time"

	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	"KeepAccount/global/nats"
	accountModel "KeepAccount/model/account"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"KeepAccount/util/timeTool"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/songzhibin97/gkit/egroup"
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

// Login
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserLogin	true	"login data"
//	@Success	200		{object}	response.Data{Data=response.Login}
//	@Router		/public/user/login [post]
func (p *PublicApi) Login(ctx *gin.Context) {
	var requestData request.UserLogin
	var err error
	// handle error
	var loginFailResponseFunc = func() {
		if err != nil {
			key := global.Cache.GetKey(constant.LoginFailCount, requestData.Email)
			count, existCache := global.Cache.GetInt(key)
			if existCache {
				if count > 5 {
					response.FailToError(ctx, errors.New("错误次数过的，请稍后再试"))
					return
				} else {
					_ = global.Cache.Increment(key, 1)
				}
			} else {
				global.Cache.Set(key, 1, time.Hour*12)
			}
			response.FailToError(ctx, err)
			return
		}
	}
	defer loginFailResponseFunc()
	// check
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		return
	}
	if false == captchaStore.Verify(requestData.CaptchaId, requestData.Captcha, true) {
		response.FailWithMessage("验证码错误", ctx)
		return
	}

	client := contextFunc.GetClient(ctx)
	// handler
	var user userModel.User
	var clientBaseInfo userModel.UserClientBaseInfo
	var responseData response.Login
	var customClaims jwt.RegisteredClaims
	user, clientBaseInfo, responseData.Token, customClaims, err = userService.Login(
		requestData.Email, requestData.Password, client, ctx,
	)
	if err != nil {
		err = errors.New("用户名不存在或者密码错误")
		return
	}
	responseData.TokenExpirationTime = customClaims.ExpiresAt.Time
	err = responseData.User.SetData(user)
	if err != nil {
		return
	}
	err = responseData.SetDataFormClientInto(clientBaseInfo)
	if err != nil {
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

// Register
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserRegister	true	"register data"
//	@Success	200		{object}	response.Data{Data=response.Login}
//	@Router		/public/user/register [post]
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

	user, err := userService.Register(data, ctx)
	if responseError(err, ctx) {
		return
	}
	// 注册成功 获取token
	customClaims := commonService.MakeCustomClaims(user.ID)
	token, err := commonService.GenerateJWT(customClaims)
	if responseError(err, ctx) {
		return
	}

	responseData := response.Register{Token: token, TokenExpirationTime: customClaims.ExpiresAt.Time}
	err = responseData.User.SetData(user)
	if responseError(err, ctx) {
		return
	}
	response.OkWithDetailed(responseData, "注册成功", ctx)
}

// TourRequest
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.TourApply	true	"data"
//	@Success	200		{object}	response.Data{Data=response.Login}
//	@Router		/public/user/tour [post]
func (p *PublicApi) TourRequest(ctx *gin.Context) {
	var requestData request.TourApply
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if !requestData.CheckSign() {
		response.Forbidden(ctx)
		return
	}
	user, err := userService.EnableTourist(requestData.DeviceNumber, contextFunc.GetClient(ctx), ctx)
	if err != nil {
		if errors.Is(err, global.ErrTooManyTourists) {
			ctx.Header("Cache-Control", "max-age=300")
		}
		response.FailToError(ctx, err)
		return
	}
	// response
	customClaims := commonService.MakeCustomClaims(user.ID)
	token, err := commonService.GenerateJWT(customClaims)
	if responseError(err, ctx) {
		return
	}
	responseData := response.Login{Token: token, TokenExpirationTime: customClaims.ExpiresAt.Time}
	err = responseData.User.SetData(user)
	if responseError(err, ctx) {
		return
	}
	clientInfo, err := user.GetUserClient(contextFunc.GetClient(ctx), db.Db)
	if responseError(err, ctx) {
		return
	}
	err = responseData.SetDataFormClientInto(clientInfo)
	if responseError(err, ctx) {
		return
	}
	response.OkWithDetailed(responseData, "welcome", ctx)
}

// UpdatePassword
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserForgetPassword	true	"data"
//	@Success	204		{object}	response.NoContent
//	@Router		/public/user/password [put]
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
	err = userService.UpdatePassword(user, requestData.Password, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// RefreshToken
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Data{Data=response.Token}
//	@Router		/user/token/refresh [post]
func (u *UserApi) RefreshToken(ctx *gin.Context) {
	claims := contextFunc.GetClaims(ctx)
	token, newClaims, err := commonService.RefreshJWT(claims)
	if responseError(err, ctx) {
		return
	}
	responseData := response.Token{Token: token, TokenExpirationTime: newClaims.ExpiresAt.Time}
	response.OkWithData(responseData, ctx)
}

// UpdatePassword
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserUpdatePassword	true	"data"
//	@Success	204		{object}	response.NoContent
//	@Router		/user/password [put]
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

	err = userService.UpdatePassword(user, requestData.Password, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// UpdateInfo
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserUpdateInfo	true	"data"
//	@Success	204		{object}	response.NoContent
//	@Router		/user/current [put]
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
	err = userService.UpdateInfo(user, requestData.Username, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// SearchUser
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserSearch	true	"data"
//	@Success	200		{object}	response.Data{Data=response.List[response.UserInfo]{}}
//	@Router		/user/search [get]
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

// SetCurrentAccount
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.Id	true	"data"
//	@Success	204		{object}	response.NoContent
//	@Router		/user/client/current/account [put]
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

	err := userService.SetClientAccount(accountUser, contextFunc.GetClient(ctx), account, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// SetCurrentShareAccount
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.Id	true	"data"
//	@Success	204		{object}	response.NoContent
//	@Router		/user/client/current/share/account [put]
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

	err := userService.SetClientShareAccount(accountUser, contextFunc.GetClient(ctx), account, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// SendCaptchaEmail
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserSendEmail	true	"data"
//	@Success	200		{object}	response.Data{Data=response.ExpirationTime}
//	@Router		/user/client/current/share/account [post]
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
	isSuccess := nats.PublishTaskWithPayload(
		nats.TaskSendCaptchaEmail, nats.PayloadSendCaptchaEmail{
			Email: user.Email, Action: requestData.Type,
		},
	)
	if !isSuccess {
		response.FailToError(ctx, errors.New("发送失败"))
	}
	response.OkWithData(response.ExpirationTime{ExpirationTime: global.Config.Captcha.EmailCaptchaTimeOut}, ctx)
}

// Home
//
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserHome	true	"data"
//	@Success	200		{object}	response.Data{Data=response.UserHome}
//	@Router		/user/home [get]
func (u *UserApi) Home(ctx *gin.Context) {
	var requestData request.UserHome
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account, _, pass := checkFunc.AccountBelongAndGet(requestData.AccountId, ctx)
	if false == pass {
		return
	}

	group := egroup.WithContext(ctx)
	nowTime, timeLocation := account.GetNowTime(), account.GetTimeLocation()
	year, month, day := nowTime.Date()
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
			IEStatistic: result,
			StartTime:   startTime,
			EndTime:     endTime,
		}
		return nil
	}
	handelGoroutineOne := func() error {
		var err error
		// 今日统计
		if err = handelOneTime(
			&todayData,
			time.Date(year, month, day, 0, 0, 0, 0, timeLocation),
			time.Date(year, month, day, 23, 59, 59, 0, timeLocation),
		); err != nil {
			return err
		}
		// 昨日统计
		if err = handelOneTime(
			&yesterdayData,
			time.Date(year, month, day-1, 0, 0, 0, 0, timeLocation),
			time.Date(year, month, day-1, 23, 59, 59, 0, timeLocation),
		); err != nil {
			return err
		}
		// 周统计
		if err = handelOneTime(
			&weekData,
			timeTool.GetFirstSecondOfWeek(nowTime),
			timeTool.GetLastSecondOfWeek(nowTime),
		); err != nil {
			return err
		}
		return err
	}

	handelGoroutineTwo := func() error {
		var err error
		// 月统计
		if err = handelOneTime(
			&monthData,
			timeTool.GetFirstSecondOfMonth(nowTime),
			timeTool.GetLastSecondOfMonth(nowTime),
		); err != nil {
			return err
		}
		// 年统计
		if err = handelOneTime(
			&yearData,
			timeTool.GetFirstSecondOfYear(nowTime),
			timeTool.GetLastSecondOfYear(nowTime),
		); err != nil {
			return err
		}
		return err
	}
	group.Go(handelGoroutineOne)
	group.Go(handelGoroutineTwo)
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

// GetTransactionShareConfig
//
//	@Tags		User/Config
//	@Produce	json
//	@Success	200	{object}	response.Data{Data=response.UserTransactionShareConfig}
//	@Router		/user/transaction/share/config [get]
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

// UpdateTransactionShareConfig
//
//	@Tags		User/Config
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserTransactionShareConfigUpdate	true	"data"
//	@Success	200		{object}	response.Data{Data=response.UserTransactionShareConfig}
//	@Router		/user/transaction/share/config [put]
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
		err = config.OpenDisplayFlag(flag, db.Db)
	} else {
		err = config.ClosedDisplayFlag(flag, db.Db)
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

// GetFriendList
//
//	@Tags		User/Friend
//	@Produce	json
//	@Success	200	{object}	response.Data{Data=response.List[response.UserInfo]{}}
//	@Router		/user/friend/list [get]
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
		CreateTime: data.CreatedAt,
	}
	responseData.Inviter.SetMaskData(inviterInfo)
	responseData.Inviter.SetMaskData(inviteeInfo)
	return
}

// CreateFriendInvitation
//
//	@Tags		User/Friend/Invitation
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserCreateFriendInvitation	true	"data"
//	@Success	200		{object}	response.Data{Data=response.UserFriendInvitation}
//	@Router		/user/friend/invitation [post]
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
	var invitee userModel.User
	invitee, err = userModel.NewDao().SelectById(requestData.Invitee)
	if responseError(err, ctx) {
		return
	}
	invitation, err = userService.Friend.CreateInvitation(user, invitee, ctx)
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

// AcceptFriendInvitation
//
//	@Tags		User/Friend/Invitation
//	@Produce	json
//	@Param		id	path		int	true	"Invitation ID"
//	@Success	200	{object}	response.Data{Data=response.UserFriendInvitation}
//	@Router		/user/friend/invitation/{id}/accept [put]
func (u *UserApi) AcceptFriendInvitation(ctx *gin.Context) {
	invitation, pass := u.getFriendInvitationByParam(ctx)
	if false == pass {
		return
	}
	if invitation.Invitee != contextFunc.GetUserId(ctx) {
		response.FailToError(ctx, errors.New("非被邀请者！"))
		return
	}
	_, _, err := userService.Friend.AcceptInvitation(&invitation, ctx)
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

// RefuseFriendInvitation
//
//	@Tags		User/Friend/Invitation
//	@Produce	json
//	@Param		id	path		int	true	"Invitation ID"
//	@Success	200	{object}	response.Data{Data=response.UserFriendInvitation}
//	@Router		/user/friend/invitation/{id}/refuse [put]
func (u *UserApi) RefuseFriendInvitation(ctx *gin.Context) {
	invitation, pass := u.getFriendInvitationByParam(ctx)
	if false == pass {
		return
	}
	if invitation.Invitee != contextFunc.GetUserId(ctx) {
		response.FailToError(ctx, errors.New("非被邀请者！"))
		return
	}
	err := userService.Friend.RefuseInvitation(&invitation, ctx)
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

// GetFriendInvitationList
//
//	@Tags		User/Friend/Invitation
//	@Produce	json
//	@Success	200	{object}	response.Data{Data=response.List[response.UserFriendInvitation]{}}
//	@Router		/user/friend/invitation [get]
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

// GetAccountInvitationList
//
//	@Tags		User/Friend/Invitation
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.UserGetAccountInvitationList	true	"query condition"
//	@Success	200		{object}	response.Data{Data=response.AccountUserInvitation}
//	@Router		/user/account/invitation/list [get]
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
