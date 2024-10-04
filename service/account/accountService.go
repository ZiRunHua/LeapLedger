package accountService

import (
	"KeepAccount/global"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"context"
	"github.com/pkg/errors"
	"strings"
)

type base struct{}

func (b *base) NewCreateData(name string, icon string, aType accountModel.Type, location string) accountModel.Account {
	return accountModel.Account{
		Type:     aType,
		Name:     strings.TrimSpace(name),
		Icon:     strings.TrimSpace(icon),
		Location: strings.TrimSpace(location),
	}
}
func (b *base) CreateOne(
	user userModel.User, data accountModel.Account, ctx context.Context,
) (account accountModel.Account, aUser accountModel.User, err error) {
	if data.Name == "" || data.Icon == "" {
		err = global.NewErrDataIsEmpty("请填写")
		return
	}
	if data.Location == "" {
		err = global.NewErrDataIsEmpty("请选择地区")
		return
	}
	account = data
	account.UserId = user.ID
	err = db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			var err error
			account, err = accountModel.NewDao(ctx.GetDb()).Create(account)
			err = db.Get(ctx).Create(&account).Error
			if err != nil {
				err = errors.WithStack(err)
			}
			aUser, err = GroupApp.Share.AddAccountUser(account, user, accountModel.UserPermissionCreator, ctx)
			if err != nil {
				err = errors.WithStack(err)
			}
			return b.updateUserCurrentAfterCreate(aUser, ctx)
		},
	)
	return
}

func (b *base) updateUserCurrentAfterCreate(accountUser accountModel.User, ctx context.Context) (err error) {
	tx := db.Get(ctx)
	user, err := userModel.NewDao(tx).SelectById(accountUser.UserId)
	if err != nil {
		return err
	}
	account := accountModel.Account{ID: accountUser.AccountId}
	err = account.ForShare(tx)
	if err != nil {
		return err
	}
	processFunc := func(clientInfo userModel.Client) error {
		updates := make(map[string]interface{})
		if clientInfo.IsCurrentShareAccount(0) && account.Type == accountModel.TypeShare {
			updates["current_share_account_id"] = account.ID
		}
		if clientInfo.IsCurrentAccount(0) && account.Type == accountModel.TypeIndependent {
			updates["current_account_id"] = account.ID
		}
		if len(updates) > 0 {
			return tx.Model(clientInfo).Where("user_id = ?", clientInfo.GetUserId()).Updates(updates).Error
		}
		return nil
	}
	err = userServer.ProcessAllClient(user, processFunc, ctx)
	if err != nil {
		return err
	}
	return
}

func (b *base) Delete(account accountModel.Account, accountUser accountModel.User, ctx context.Context) (err error) {
	tx := db.Get(ctx)
	if accountUser.AccountId != account.ID {
		panic("err accountId")
	}
	if false == accountUser.HavePermission(accountModel.UserPermissionEditAccount) {
		return global.ErrNoPermission
	}
	err = tx.Delete(&account).Error
	if err != nil {
		return err
	}
	err = tx.Delete(&accountModel.User{}, "account_id = ?", account.ID).Error
	if err != nil {
		return err
	}
	err = tx.Delete(&accountModel.Mapping{}, "main_id = ? or related_id = ?", account.ID, account.ID).Error
	if err != nil {
		return err
	}
	// 删除的可能是当前账本 故需要更新客户端信息
	err = b.updateUserCurrentAfterDelete(accountUser, ctx)
	if err != nil {
		return err
	}
	return
}

func (b *base) updateUserCurrentAfterDelete(accountUser accountModel.User, ctx context.Context) (err error) {
	tx := db.Get(ctx)
	user, err := userModel.NewDao().SelectById(accountUser.UserId)
	if err != nil {
		return err
	}
	account, shareAccount, err := b.getNewCurrentAccount(user, ctx)
	if err != nil {
		return err
	}
	processFunc := func(clientInfo userModel.Client) error {
		updates := make(map[string]interface{})
		if clientInfo.IsCurrentShareAccount(accountUser.AccountId) {
			if shareAccount == nil {
				updates["current_share_account_id"] = 0
			} else {
				updates["current_share_account_id"] = shareAccount.ID
			}
		}
		if clientInfo.IsCurrentAccount(accountUser.AccountId) {
			updates["current_account_id"] = account.ID
		}
		if len(updates) > 0 {
			return tx.Model(clientInfo).Where("user_id = ?", clientInfo.GetUserId()).Updates(updates).Error
		}
		return nil
	}
	err = userServer.ProcessAllClient(user, processFunc, ctx)
	if err != nil {
		return err
	}
	return
}

func (b *base) getNewCurrentAccount(user userModel.User, ctx context.Context) (
	accountModel.Account, *accountModel.Account, error,
) {
	tx := db.Get(ctx)
	dao, condition := accountModel.NewDao(tx), *accountModel.NewUserCondition()
	current, err := dao.SelectByUserAndAccountType(user.ID, condition)
	if err != nil {
		return current, nil, err
	}

	var currentShare *accountModel.Account
	condition.SetType(accountModel.TypeShare)
	if account, err := dao.SelectByUserAndAccountType(user.ID, condition); err != nil {
		currentShare = &account
	}
	return current, currentShare, err
}
func (b *base) Update(
	account accountModel.Account, accountUser accountModel.User, updateData accountModel.AccountUpdateData,
	ctx context.Context,
) (err error) {
	if accountUser.AccountId != account.ID {
		return global.ErrAccountId
	}
	if false == accountUser.HavePermission(accountModel.UserPermissionEditAccount) {
		return global.ErrNoPermission
	}
	return db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			return accountModel.NewDao(ctx.GetDb()).Update(account, updateData)
		},
	)
}

func (b *base) UpdateUser(
	accountUser accountModel.User, operator accountModel.User, updateData accountModel.UserUpdateData,
	ctx context.Context,
) (result accountModel.User, err error) {
	if accountUser.AccountId != operator.AccountId {
		return result, global.ErrAccountId
	}
	if false == operator.HavePermission(accountModel.UserPermissionEditUser) {
		return result, global.ErrNoPermission
	}
	result, err = accountModel.NewDao(db.Get(ctx)).UpdateUser(accountUser, updateData)
	return
}
