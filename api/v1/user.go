package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	"KeepAccount/model/common/query"
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

func (u *PublicApi) Login(ctx *gin.Context) {
	var requestData request.UserLogin
	var err error
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

	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		return
	}

	if false == captchaStore.Verify(requestData.CaptchaId, requestData.Captcha, true) {
		response.FailWithMessage("验证码错误", ctx)
		return
	}

	client := contextFunc.GetClient(ctx)
	var currentAccount *accountModel.Account
	var token string
	var user *userModel.User
	transactionFunc := func(tx *gorm.DB) error {
		var clientBaseInfo *userModel.UserClientBaseInfo
		user, clientBaseInfo, token, err = userService.Login(requestData.Email, requestData.Password, client, tx)
		if err != nil {
			return err
		}
		if clientBaseInfo.CurrentAccountID != 0 {
			currentAccount, err = query.FirstByPrimaryKey[*accountModel.Account](clientBaseInfo.CurrentAccountID)
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
	if token == "" {
		err = errors.New("token获取失败")
		return
	}
	if err == nil {
		response.OkWithDetailed(
			response.Login{
				Token: token, CurrentAccount: response.AccountModelToResponse(currentAccount),
				User: response.UserModelToResponse(user),
			}, "登录成功", ctx,
		)
	}
}

func (u *PublicApi) Register(ctx *gin.Context) {
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

	data := &userModel.AddData{Username: requestData.Username, Password: requestData.Password, Email: requestData.Email}
	var user *userModel.User
	var token string
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			user, err = userService.Register(data, tx)
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
	_ = thirdpartyService.SendNotificationEmail(constant.NotificationOfRegistrationSuccess, user)
	response.OkWithDetailed(
		response.Register{
			Token: token,
		}, "注册成功", ctx,
	)
}

func (u *PublicApi) UpdatePassword(ctx *gin.Context) {
	var requestData request.UserForgetPassword
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	err := commonService.CheckEmailCaptcha(requestData.Email, requestData.Captcha)
	if responseError(err, ctx) {
		return
	}
	user, err := userModel.Dao.NewUser(nil).SelectByEmail(requestData.Email)
	if responseError(err, ctx) {
		return
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return userService.UpdatePassword(user, requestData.Password, tx)
		},
	)
	// 发送不成功不影响主流程
	_ = thirdpartyService.SendNotificationEmail(constant.NotificationOfUpdatePassword, user)
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
			return userService.UpdatePassword(user, requestData.Password, tx)
		},
	)
	// 发送不成功不影响主流程
	_ = thirdpartyService.SendNotificationEmail(constant.NotificationOfUpdatePassword, user)
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
	var user *userModel.User
	user, err = contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	err = global.GvaDb.Model(user).Update("username", requestData.Username).Error
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
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

	group := egroup.WithContext(ctx)
	nowTime := time.Now()
	year, month, day := time.Now().Date()
	var todayData, yesterdayData, weekData, monthData, yearData response.TransactionStatistic
	condition := &transactionModel.StatisticCondition{
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{AccountId: &requestData.AccountId},
	}
	handelOneTime := func(data *response.TransactionStatistic, start time.Time, end time.Time) error {
		result, err := transactionModel.Dao.NewTransaction(nil).GetStatisticByCondition(
			condition,
			start,
			end,
		)
		if err != nil {
			return err
		}
		*data = response.TransactionStatistic{
			IncomeExpenseStatistic: *result,
			StartTime:              start.Unix(),
			EndTime:                end.Unix(),
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
	responseData.HeaderCard = &response.UserHomeHeaderCard{&monthData}
	responseData.TimePeriodStatistics = &response.UserHomeTimePeriodStatistics{
		TodayData: &todayData, YesterdayData: &yesterdayData, WeekData: &weekData, YearData: &yearData,
	}
	response.OkWithData(responseData, ctx)
}
