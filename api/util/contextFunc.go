package util

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global/constant"
	"KeepAccount/global/cusCtx"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"github.com/gin-gonic/gin"
	"strconv"
)

var ContextFunc = new(contextFunc)

type contextFunc struct{}

func (cf *contextFunc) GetToken(ctx *gin.Context) string {
	return ctx.Request.Header.Get("authorization")
}

func (cf *contextFunc) SetClaims(claims *util.CustomClaims, ctx *gin.Context) {
	ctx.Set(string(cusCtx.Claims), claims)
}

func (cf *contextFunc) GetClaims(ctx *gin.Context) util.CustomClaims {
	return ctx.MustGet(string(cusCtx.Claims)).(util.CustomClaims)
}

func (cf *contextFunc) SetUserId(id uint, ctx *gin.Context) {
	ctx.Set(string(cusCtx.UserId), id)
}

func (cf *contextFunc) GetUserId(ctx *gin.Context) uint {
	return ctx.MustGet(string(cusCtx.UserId)).(uint)
}

func (cf *contextFunc) GetUser(ctx *gin.Context) (userModel.User, error) {
	value, exits := ctx.Get(string(cusCtx.UserId))
	if exits {
		return value.(userModel.User), nil
	}
	var user userModel.User
	err := db.Db.First(&user, cf.GetUserId(ctx)).Error
	ctx.Set(string(cusCtx.UserId), user)
	return user, err
}

func (cf *contextFunc) SetAccountId(id uint, ctx *gin.Context) {
	ctx.Set(string(cusCtx.AccountId), id)
}

func (cf *contextFunc) GetAccountId(ctx *gin.Context) uint {
	return ctx.MustGet(string(cusCtx.AccountId)).(uint)
}

func (cf *contextFunc) GetAccount(ctx *gin.Context) accountModel.Account {
	value, exist := ctx.Get(string(cusCtx.Account))
	if exist {
		return value.(accountModel.Account)
	}
	account, accountUser, exist := cf.GetAccountByParam(ctx, false)
	if !exist {
		panic("account not exist")
	}
	ctx.Set(string(cusCtx.Account), account)
	ctx.Set(string(cusCtx.AccountUser), accountUser)
	return account
}

func (cf *contextFunc) GetAccountUser(ctx *gin.Context) accountModel.User {
	value, exist := ctx.Get(string(cusCtx.AccountUser))
	if exist {
		return value.(accountModel.User)
	}
	account, accountUser, exist := cf.GetAccountByParam(ctx, false)
	if !exist {
		panic("account not exist")
	}
	ctx.Set(string(cusCtx.Account), account)
	ctx.Set(string(cusCtx.AccountUser), accountUser)
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
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		panic(err)
	}
	return uint(id)
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
	return cf.GetUintParamByKey(string(cusCtx.AccountId), ctx)
}

func (cf *contextFunc) CheckAccountPermissionFromParam(ctx *gin.Context, permission accountModel.UserPermission) (pass bool) {
	accountId, pass := cf.GetAccountIdByParam(ctx)
	if !pass {
		return
	}
	pass, err := accountModel.NewDao().CheckUserPermission(permission, accountId, cf.GetUserId(ctx))
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if !pass {
		response.Forbidden(ctx)
	}
	return true
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

func (cf *contextFunc) GetCacheKey(t constant.CacheTab, ctx *gin.Context) string {
	return string(t) + "_" + ctx.ClientIP()
}
