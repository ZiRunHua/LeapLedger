package util

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"github.com/gin-gonic/gin"
	"strconv"
)

var ContextFunc = new(contextFunc)

const (
	_UserId = "_user_id_"
	_Claims = "_claims_"
)

type contextFunc struct {
}

func (cf *contextFunc) SetUserId(id uint, ctx *gin.Context) {
	ctx.Set(_UserId, id)
}

func (cf *contextFunc) SetClaims(claims *util.CustomClaims, ctx *gin.Context) {
	ctx.Set(_Claims, claims)
}

func (cf *contextFunc) GetUserId(ctx *gin.Context) uint {
	return ctx.MustGet(_UserId).(uint)
}

func (cf *contextFunc) GetUser(ctx *gin.Context) (userModel.User, error) {
	user := new(userModel.User)
	err := global.GvaDb.First(user, cf.GetUserId(ctx)).Error
	return *user, err
}

func (cf *contextFunc) GetToken(ctx *gin.Context) string {
	return ctx.Request.Header.Get("authorization")
}

func (cf *contextFunc) GetClaims(ctx *gin.Context) util.CustomClaims {
	return ctx.MustGet(_Claims).(util.CustomClaims)
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

func (cf *contextFunc) GetUintParamByKey(key string, ctx *gin.Context) (uint, bool) {
	id, err := strconv.Atoi(ctx.Param(key))
	if err != nil {
		response.FailToParameter(ctx, err)
		return 0, false
	}
	return uint(id), true
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
