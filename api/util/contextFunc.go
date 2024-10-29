package util

import (
	"github.com/ZiRunHua/LeapLedger/api/request"
	"github.com/ZiRunHua/LeapLedger/api/response"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/global/cus"
	"github.com/ZiRunHua/LeapLedger/global/db"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

var ContextFunc = new(contextFunc)

type contextFunc struct{}

func (cf *contextFunc) GetToken(ctx *gin.Context) string {
	return ctx.Request.Header.Get("authorization")
}

func (cf *contextFunc) SetClaims(claims jwt.RegisteredClaims, ctx *gin.Context) {
	ctx.Set(string(cus.Claims), claims)
}

func (cf *contextFunc) GetClaims(ctx *gin.Context) jwt.RegisteredClaims {
	return ctx.Value(string(cus.Claims)).(jwt.RegisteredClaims)
}

func (cf *contextFunc) SetUserId(id uint, ctx *gin.Context) {
	ctx.Set(string(cus.UserId), id)
}

func (cf *contextFunc) GetUserId(ctx *gin.Context) uint {
	return cf.GetUint(cus.UserId, ctx)
}

func (cf *contextFunc) GetUser(ctx *gin.Context) (userModel.User, error) {
	value, exits := ctx.Get(string(cus.User))
	if exits {
		return value.(userModel.User), nil
	}
	var user userModel.User
	err := db.Db.First(&user, cf.GetUserId(ctx)).Error
	ctx.Set(string(cus.User), user)
	return user, err
}

func (cf *contextFunc) SetAccountId(id uint, ctx *gin.Context) {
	ctx.Set(string(cus.AccountId), id)
}

func (cf *contextFunc) GetAccountId(ctx *gin.Context) uint {
	return cf.GetUint(cus.AccountId, ctx)
}

func (cf *contextFunc) GetAccount(ctx *gin.Context) accountModel.Account {
	value, exist := ctx.Get(string(cus.Account))
	if exist {
		return value.(accountModel.Account)
	}
	account, accountUser, exist := cf.GetAccountByParam(ctx, false)
	if !exist {
		panic("account not exist")
	}
	ctx.Set(string(cus.Account), account)
	ctx.Set(string(cus.AccountUser), accountUser)
	return account
}

func (cf *contextFunc) GetAccountUser(ctx *gin.Context) accountModel.User {
	value, exist := ctx.Get(string(cus.AccountUser))
	if exist {
		return value.(accountModel.User)
	}
	account, accountUser, exist := cf.GetAccountByParam(ctx, false)
	if !exist {
		panic("account not exist")
	}
	ctx.Set(string(cus.Account), account)
	ctx.Set(string(cus.AccountUser), accountUser)
	return accountUser
}

func (cf *contextFunc) GetClient(ctx *gin.Context) constant.Client {
	userAgent := ctx.GetHeader("User-Agent")
	for clientType, client := range userModel.GetClients() {
		if client.CheckUserAgent(userAgent) {
			return clientType
		}
	}
	panic("Not found client")
}

func (cf *contextFunc) GetUserCurrentClientInfo(ctx *gin.Context) (userModel.UserClientBaseInfo, error) {
	return userModel.NewDao().SelectUserClientBaseInfo(cf.GetUserId(ctx), cf.GetClient(ctx))
}

func (cf *contextFunc) GetId(ctx *gin.Context) uint {
	return cf.GetUint("id", ctx)
}

func (cf *contextFunc) GetUintParamByKey(key string, ctx *gin.Context) (uint, bool) {
	id, err := strconv.Atoi(ctx.Param(key))
	if err != nil {
		response.FailToParameter(ctx, err)
		return 0, false
	}
	return uint(id), true
}

func (cf *contextFunc) GetAccountIdByParam(ctx *gin.Context) (uint, bool) {
	return cf.GetUintParamByKey(string(cus.AccountId), ctx)
}

func (cf *contextFunc) CheckAccountPermissionFromParam(
	permission accountModel.UserPermission, ctx *gin.Context) (pass bool) {
	accountId, pass := cf.GetAccountIdByParam(ctx)
	if !pass {
		return
	}
	return CheckFunc.AccountPermission(accountId, permission, ctx)
}

func (cf *contextFunc) GetInfoTypeFormParam(ctx *gin.Context) request.InfoType {
	return request.InfoType(ctx.Param("type"))
}

func (cf *contextFunc) GetAccountType(ctx *gin.Context) accountModel.Type {
	return accountModel.Type(ctx.Param("type"))
}

func (cf *contextFunc) GetParamId(ctx *gin.Context) (uint, bool) {
	return cf.GetUintParamByKey("id", ctx)
}

func (cf *contextFunc) GetTimeLocation(ctx *gin.Context) *time.Location {
	l, err := time.LoadLocation(cf.GetAccount(ctx).Location)
	if err != nil {
		panic(err)
	}
	return l
}

func (cf *contextFunc) GetNowTime(ctx *gin.Context) time.Time {
	account := cf.GetAccount(ctx)
	return account.GetNowTime()
}

func (cf *contextFunc) GetUint(key cus.Key, ctx *gin.Context) uint {
	param := ctx.Param(string(key))
	if len(param) != 0 {
		id, err := strconv.Atoi(param)
		if err != nil {
			panic(err)
		}
		return uint(id)
	}
	switch value := ctx.Value(string(key)).(type) {
	case uint:
		return value
	case int:
		return uint(value)
	case string:
		id, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		return uint(id)
	default:
		return value.(uint)
	}
}
