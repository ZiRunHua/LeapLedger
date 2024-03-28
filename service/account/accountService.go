package accountService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type base struct{}

func (b *base) CreateOne(
	user userModel.User, name string, icon string, aType accountModel.Type, tx *gorm.DB,
) (account accountModel.Account, aUser accountModel.User, err error) {
	if name == "" || icon == "" {
		err = global.NewErrDataIsEmpty("name或icon")
		return
	}
	account = accountModel.Account{
		UserId: user.ID,
		Name:   name,
		Icon:   icon,
		Type:   aType,
	}
	err = tx.Create(&account).Error
	if err != nil {
		err = errors.WithStack(err)
	}
	aUser, err = ServiceGroupApp.Share.AddAccountUser(account, user, accountModel.UserPermissionCreator, tx)
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

func (b *base) Delete(account accountModel.Account, accountUser accountModel.User, tx *gorm.DB) (err error) {
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
	//删除的可能是当前账本 故需要更新客户端信息
	err = b.updateUserCurrentAfterDelete(accountUser, tx)
	if err != nil {
		return err
	}
	return
}

func (b *base) updateUserCurrentAfterDelete(
	accountUser accountModel.User, tx *gorm.DB,
) (err error) {
	var newCurrentAccount, newCurrentShareAccount accountModel.Account
	dao := accountModel.NewDao(tx)
	condition := *accountModel.NewUserCondition()
	newCurrentAccount, err = dao.SelectByUserAndAccountType(accountUser.UserId, condition)
	if err != nil && false == errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	newCurrentShareAccount, err = dao.SelectByUserAndAccountType(
		accountUser.UserId, *condition.SetType(accountModel.TypeShare),
	)
	if err != nil && false == errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var list []userModel.UserClientBaseInfo
	for _, client := range constant.ClientList {
		list, err = userModel.NewDao().SelectClientInfoByUserAndAccount(
			client, accountUser.UserId, accountUser.AccountId,
		)
		if err != nil {
			return err
		}
		for _, info := range list {
			updates := make(map[string]interface{})
			if info.CurrentShareAccountId == accountUser.AccountId {
				updates["current_share_Account_id"] = newCurrentShareAccount.ID
			}
			if info.CurrentAccountId == accountUser.AccountId {
				updates["current_Account_id"] = newCurrentAccount.ID
			}
			if len(updates) > 0 {
				err = tx.Where("user_id = ?", info.UserId).Updates(updates).Error
				if err != nil {
					return err
				}
			}
		}
	}
	return
}

func (b *base) Update(
	account accountModel.Account, accountUser accountModel.User, updateData accountModel.AccountUpdateData, tx *gorm.DB,
) (err error) {
	if accountUser.AccountId != account.ID {
		return global.ErrAccountId
	}
	if false == accountUser.HavePermission(accountModel.UserPermissionEditAccount) {
		return global.ErrNoPermission
	}
	err = accountModel.NewDao(tx).Update(account, updateData)
	if err != nil {
		return err
	}
	//删除的可能是当前账本 故需要更新客户端信息
	err = b.updateUserCurrentAfterDelete(accountUser, tx)
	if err != nil {
		return err
	}
	return
}

func (b *base) UpdateUser(
	accountUser accountModel.User, operator accountModel.User, updateData accountModel.UserUpdateData, tx *gorm.DB,
) (result accountModel.User, err error) {
	if accountUser.AccountId != operator.AccountId {
		return result, global.ErrAccountId
	}
	if false == operator.HavePermission(accountModel.UserPermissionEditUser) {
		return result, global.ErrNoPermission
	}
	result, err = accountModel.NewDao(tx).UpdateUser(accountUser, updateData)
	return
}
