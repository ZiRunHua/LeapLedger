package accountService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	logModel "KeepAccount/model/log"
	userModel "KeepAccount/model/user"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type share struct{}

func (b *share) CreateUserInvitation(
	account accountModel.Account, inviter accountModel.User, invitee userModel.User,
	ptoPermission *accountModel.UserPermission, tx *gorm.DB,
) (invitation accountModel.UserInvitation, err error) {
	if false == inviter.HavePermission(accountModel.UserPermissionInvite) {
		return invitation, global.ErrNoPermission
	}

	accountDao := accountModel.NewDao(tx)
	var existUser bool
	if existUser, err = accountDao.ExistUser(account.ID, invitee.ID); err != nil {
		return
	}
	if existUser {
		return invitation, errors.New("该用户已加入")
	}

	var permission = accountModel.UserPermissionOwnEditor
	if ptoPermission != nil {
		permission = *ptoPermission
	}

	invitation, err = accountDao.CreateUserInvitation(account.ID, inviter.UserId, invitee.ID, permission)
	if err != nil {
		if false == errors.Is(err, gorm.ErrDuplicatedKey) {
			return
		}
		// 已存在邀请则修改邀请
		invitation, err = accountDao.SelectUserInvitation(account.ID, invitee.ID)
		if err != nil {
			return
		}
		if err = invitation.ForUpdate(tx); err != nil {
			return
		}
		if invitation.Status != accountModel.UserInvitationStatsOfWaiting {
			err = invitation.Updates(inviter.UserId, permission, accountModel.UserInvitationStatsOfWaiting, tx)
			if err != nil {
				return
			}
		}
	}
	_, err = userModel.NewDao(tx).AddFriend(inviter.UserId, invitee.ID, userModel.FriendAddModeOfAccountInvitation)
	if err != nil {
		return
	}
	return
}

func (b *share) AddAccountUser(
	account accountModel.Account, user userModel.User, permission accountModel.UserPermission, tx *gorm.DB,
) (accountModel.User, error) {
	return accountModel.NewDao(tx).CreateUser(account.ID, user.ID, permission)
}

func (b *share) CheckAccountPermission(
	account accountModel.Account, user userModel.User, permission accountModel.UserPermission,
) (accountModel.User, error) {
	accountUser, err := accountModel.NewDao().SelectUser(account.ID, user.ID)
	if err != nil {
		return accountUser, err
	}
	if false == accountUser.HavePermission(permission) {
		return accountUser, global.ErrNoPermission
	}
	return accountUser, err
}

func (b *share) MappingAccount(
	user userModel.User,
	mainAccount accountModel.Account, mappingAccount accountModel.Account, tx *gorm.DB,
) (mapping accountModel.Mapping, err error) {
	_, err = b.CheckAccountPermission(mainAccount, user, accountModel.UserPermissionEditOwn)
	if err != nil {
		return
	}
	if user.ID != mappingAccount.UserId {
		return mapping, global.ErrNoPermission
	}
	if err = mainAccount.ForUpdate(tx); err != nil {
		return
	}
	if mainAccount.Type != accountModel.TypeShare {
		return mapping, errors.New("账本类型错误")
	}
	if mainAccount.ID == mappingAccount.ID {
		return mapping, errors.New("数据异常")
	}
	mapping, err = accountModel.NewDao(tx).CreateMapping(mainAccount, mappingAccount)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = errors.WithStack(errors.New("请勿重复关联"))
		}
		return
	}
	_, _, err = logServer.RecordAccountLog(
		&mapping,
		logModel.BaseAccountLog{UserId: user.ID, AccountId: mainAccount.ID, Operation: constant.LogOperationOfAdd}, tx,
	)
	if err != nil {
		return
	}
	return
}

func (b *share) DeleteAccountMapping(user userModel.User, mapping accountModel.Mapping, tx *gorm.DB) (err error) {
	_, _, err = logServer.RecordAccountLog(
		&mapping,
		logModel.BaseAccountLog{UserId: user.ID, AccountId: mapping.MainId, Operation: constant.LogOperationOfDelete},
		tx,
	)
	if err != nil {
		return err
	}
	err = accountModel.NewDao(tx).DeleteMapping(mapping)
	if err != nil {
		return
	}
	return
}

func (b *share) UpdateAccountMapping(
	user userModel.User, mapping accountModel.Mapping, mappingAccount accountModel.Account, tx *gorm.DB,
) (err error) {
	_, _, err = logServer.RecordAccountLog(
		&mapping,
		logModel.BaseAccountLog{UserId: user.ID, AccountId: mapping.MainId, Operation: constant.LogOperationOfUpdate},
		tx,
	)
	if err != nil {
		return err
	}
	err = accountModel.NewDao(tx).UpdateRelatedAccount(mapping, mappingAccount)
	if err != nil {
		return
	}
	return
}
