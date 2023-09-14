package util

import (
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type checkFunc struct {
}

var CheckFunc = new(checkFunc)

func (ckf *checkFunc) AccountBelong(id uint, ctx *gin.Context) bool {
	exist, err := accountModel.NewDao().ExistUser(id, ContextFunc.GetUserId(ctx))
	if err != nil {
		response.FailToError(ctx, err)
		return false
	}
	if false == exist {
		response.Forbidden(ctx)
		return false
	}
	return true
}

func (ckf *checkFunc) AccountBelongAndGet(accountId uint, ctx *gin.Context) (
	account accountModel.Account, accountUser accountModel.User, pass bool,
) {
	err := global.GvaDb.First(&account, accountId).Error
	if ckf.handelForbiddenOrError(err, ctx) {
		return
	}
	accountUser, err = accountModel.NewDao().SelectUser(accountId, ContextFunc.GetUserId(ctx))
	if ckf.handelForbiddenOrError(err, ctx) {
		return
	}
	return account, accountUser, true
}

func (ckf *checkFunc) TransactionCategoryBelongAndGet(id interface{}, ctx *gin.Context) (
	pass bool, category categoryModel.Category, account accountModel.Account,
) {
	err := global.GvaDb.First(&category, id).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	account, err = category.GetAccount()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if account.UserId != ContextFunc.GetUserId(ctx) {
		response.Forbidden(ctx)
		return
	}
	return true, category, account
}

func (ckf *checkFunc) FriendInvitationBelongAndGet(id interface{}, ctx *gin.Context) (
	pass bool, data userModel.FriendInvitation,
) {
	err := global.GvaDb.First(&data, id).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	currentUserId := ContextFunc.GetUserId(ctx)
	if data.Inviter != currentUserId && data.Invitee != currentUserId {
		response.Forbidden(ctx)
		return
	}
	return true, data
}
func (ckf *checkFunc) handelForbiddenOrError(err error, ctx *gin.Context) (pass bool) {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Forbidden(ctx)
			return true
		}
		response.FailToError(ctx, err)
		return true
	}
	return false
}
