package util

import (
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type checkFunc struct{}

var CheckFunc Checker = new(checkFunc)

type Checker interface {
	AccountBelong(id uint, ctx *gin.Context) bool
	AccountBelongAndGet(accountId uint, ctx *gin.Context) (accountModel.Account, accountModel.User, bool)
	AccountPermission(accountId uint, permission accountModel.UserPermission, ctx *gin.Context) bool
	CategoryBelongAndGet(categoryId uint, accountId uint, ctx *gin.Context) (categoryModel.Category, bool)
	CategoryFatherBelongAndGet(fatherId uint, accountId uint, ctx *gin.Context) (categoryModel.Father, bool)
	TransactionCategoryBelongAndGet(id interface{}, ctx *gin.Context) (bool, categoryModel.Category, accountModel.Account)
	FriendInvitationBelongAndGet(id interface{}, ctx *gin.Context) (bool, userModel.FriendInvitation)
}

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
	err := db.Db.First(&account, accountId).Error
	if ckf.handelForbiddenOrError(err, ctx) {
		return
	}
	accountUser, err = accountModel.NewDao().SelectUser(accountId, ContextFunc.GetUserId(ctx))
	if ckf.handelForbiddenOrError(err, ctx) {
		return
	}
	return account, accountUser, true
}

func (ckf *checkFunc) AccountPermission(accountId uint, permission accountModel.UserPermission, ctx *gin.Context) bool {
	pass, err := accountModel.NewDao().CheckUserPermission(permission, accountId, ContextFunc.GetUserId(ctx))
	if err != nil {
		response.FailToError(ctx, err)
		return false
	}
	if !pass {
		response.Forbidden(ctx)
	}
	return pass
}

// CategoryBelongAndGet
// Check whether the category belongs to account
func (ckf *checkFunc) CategoryBelongAndGet(categoryId uint, accountId uint, ctx *gin.Context) (category categoryModel.Category, pass bool) {
	err := db.Db.First(&category, categoryId).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if category.AccountId != accountId {
		response.FailToError(ctx, global.ErrAccountId)
		return
	}
	return category, true
}

// CategoryFatherBelongAndGet
// Check whether the category father belongs to account
func (ckf *checkFunc) CategoryFatherBelongAndGet(fatherId uint, accountId uint, ctx *gin.Context) (father categoryModel.Father, pass bool) {
	err := db.Db.First(&father, fatherId).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if father.AccountId != accountId {
		response.FailToError(ctx, global.ErrAccountId)
		return
	}
	return father, true
}

func (ckf *checkFunc) TransactionCategoryBelongAndGet(id interface{}, ctx *gin.Context) (
	pass bool, category categoryModel.Category, account accountModel.Account,
) {
	err := db.Db.First(&category, id).Error
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
	err := db.Db.First(&data, id).Error
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
