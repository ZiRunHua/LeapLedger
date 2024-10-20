package accountModel

import (
	"KeepAccount/global"
	userModel "KeepAccount/model/user"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type User struct {
	ID         uint `gorm:"primarykey"`
	AccountId  uint `gorm:"not null;uniqueIndex:idx_mapping,priority:1"`
	UserId     uint `gorm:"not null;uniqueIndex:idx_mapping,priority:2"`
	Permission UserPermission
	CreatedAt  time.Time      `gorm:"type:TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"type:TIMESTAMP"`
	DeletedAt  gorm.DeletedAt `gorm:"index;type:TIMESTAMP"`
}

type UserUpdateData struct {
	Permission UserPermission
}

// UserPermission 用户权限
type UserPermission uint

const (
	UserPermissionAddOwn UserPermission = 1 << iota
	UserPermissionEditOwn
	UserPermissionReadOwn
	UserPermissionAddOther
	UserPermissionEditOther
	UserPermissionReadOther
	UserPermissionEditUser
	UserPermissionEditAccount
	UserPermissionInvite
)

const UserPermissionReader = UserPermissionReadOther + UserPermissionReadOwn
const UserPermissionOwnEditor = UserPermissionReader + UserPermissionAddOwn + UserPermissionEditOwn + UserPermissionInvite
const UserPermissionAdministrator = UserPermissionOwnEditor + UserPermissionEditOther
const UserPermissionCreator = UserPermissionOwnEditor + UserPermissionEditUser + UserPermissionEditAccount

func (up *UserPermission) ToRole() UserRole {
	switch *up {
	case UserPermissionReader:
		return Reader
	case UserPermissionOwnEditor:
		return OwnEditor
	case UserPermissionAdministrator:
		return Administrator
	case UserPermissionCreator:
		return Creator
	default:
		panic("不存在该权限角色")
	}
}

// UserRole 用户角色
type UserRole string

const Reader UserRole = "reader"
const OwnEditor UserRole = "ownEditor"
const Administrator UserRole = "administrator"
const Creator UserRole = "creator"

func (ur *UserRole) ToUserPermission() UserPermission {
	switch *ur {
	case Reader:
		return UserPermissionReader
	case OwnEditor:
		return UserPermissionOwnEditor
	case Administrator:
		return UserPermissionAdministrator
	case Creator:
		return UserPermissionCreator
	default:
		panic("不存该角色")
	}
}

func (u *User) TableName() string {
	return "account_user"
}

func (u *User) SelectById(id uint) error {
	return global.GvaDb.First(u, id).Error
}

func (u *User) HavePermission(permission UserPermission) bool {
	return (u.Permission & permission) > 0
}

func (u *User) CheckTransEditByUserId(transOwner uint) error {
	var pass bool
	if transOwner == u.UserId {
		pass = u.HavePermission(UserPermissionEditOwn)
	} else {
		pass = u.HavePermission(UserPermissionEditOther)
	}
	if false == pass {
		return global.ErrNoPermission
	}
	return nil
}

func (u *User) CheckTransAddByUserId(transOwner uint) error {
	var pass bool
	if transOwner == u.UserId {
		pass = u.HavePermission(UserPermissionAddOwn)
	} else {
		pass = u.HavePermission(UserPermissionAddOther)
	}
	if false == pass {
		return global.ErrNoPermission
	}
	return nil
}

func (u *User) GetAccount() (account Account, err error) {
	err = account.SelectById(u.AccountId)
	return
}

func (u *User) GetConfig() (config UserConfig, err error) {
	return NewDao().SelectUserConfig(u.AccountId, u.UserId)
}

func (u *User) GetUserInfo() (userModel.UserInfo, error) {
	return userModel.NewDao().SelectUserInfoById(u.UserId)
}

func (u *User) GetRole() UserRole {
	return u.Permission.ToRole()
}

type UserInvitation struct {
	ID         uint `gorm:"primarykey"`
	AccountId  uint `gorm:"uniqueIndex:idx_mapping,priority:1"`
	Inviter    uint
	Invitee    uint `gorm:"uniqueIndex:idx_mapping,priority:2"`
	Status     UserInvitationStatus
	Permission UserPermission
	CreatedAt  time.Time      `gorm:"type:TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"type:TIMESTAMP"`
	DeletedAt  gorm.DeletedAt `gorm:"index;type:TIMESTAMP"`
}

type UserInvitationStatus int

const (
	UserInvitationStatsOfWaiting UserInvitationStatus = iota
	UserInvitationStatsOfAccept
	UserInvitationStatsOfRefuse
)

func (u *UserInvitation) TableName() string {
	return "account_user_invitation"
}

func (u *UserInvitation) ForUpdate(tx *gorm.DB) error {
	if err := tx.Model(u).Clauses(clause.Locking{Strength: "UPDATE"}).First(u, u.ID).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserInvitation) ForShare(tx *gorm.DB) error {
	if err := tx.Model(u).Clauses(clause.Locking{Strength: "SHARE"}).First(u, u.ID).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserInvitation) GetAccount() (Account, error) {
	return NewDao().SelectById(u.AccountId)
}

func (u *UserInvitation) GetInviterInfo() (userModel.UserInfo, error) {
	return userModel.NewDao().SelectUserInfoById(u.Inviter)
}

func (u *UserInvitation) GetInviteeInfo() (userModel.UserInfo, error) {
	return userModel.NewDao().SelectUserInfoById(u.Invitee)
}

func (u *UserInvitation) GetRole() UserRole {
	return u.Permission.ToRole()
}

func (u *UserInvitation) Accept(tx *gorm.DB) (user User, err error) {
	err = u.ForShare(tx)
	if err != nil {
		return
	}
	if u.Status == UserInvitationStatsOfAccept {
		user, err = NewDao(tx).SelectUser(u.AccountId, u.Invitee)
		return
	} else if u.Status == UserInvitationStatsOfRefuse {
		err = errors.New("邀请状态异常")
		return
	}

	err = u.UpdateStatus(UserInvitationStatsOfAccept, tx)
	return
}

func (u *UserInvitation) Refuse(tx *gorm.DB) (err error) {
	err = u.ForShare(tx)
	if err != nil {
		return
	}
	if u.Status == UserInvitationStatsOfRefuse {
		return
	} else if u.Status == UserInvitationStatsOfAccept {
		err = errors.New("邀请状态异常")
		return
	}

	err = u.UpdateStatus(UserInvitationStatsOfRefuse, tx)
	if err != nil {
		return
	}
	return
}

func (u *UserInvitation) UpdateStatus(status UserInvitationStatus, tx *gorm.DB) error {
	err := tx.Model(u).Update("status", status).Error
	if err != nil {
		u.Status = status
	}
	return err
}

func (u *UserInvitation) Updates(
	inviterId uint, permission UserPermission, status UserInvitationStatus, tx *gorm.DB,
) error {
	err := tx.Model(u).Where("id = ?", u.ID).Select("inviter", "permission", "status").Updates(
		UserInvitation{
			Inviter:    inviterId,
			Permission: permission,
			Status:     status,
		},
	).Error
	return err
}

type UserConfig struct {
	ID         uint           `gorm:"primarykey"`
	AccountId  uint           `gorm:"not null;uniqueIndex:idx_mapping,priority:1"`
	UserId     uint           `gorm:"not null;uniqueIndex:idx_mapping,priority:2"`
	TransFlags TransFlag      `gorm:"type:smallint;unsigned;comment:'交易配置标志'"`
	CreatedAt  time.Time      `gorm:"type:TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"type:TIMESTAMP"`
	DeletedAt  gorm.DeletedAt `gorm:"index;type:TIMESTAMP"`
}

func (uc *UserConfig) TableName() string {
	return "account_user_config"
}

type UserConfigFlag uint
type TransFlag UserConfigFlag

const (
	Flag_Trans_Sync_Mapping_Account TransFlag = 1 << iota
)
const DefaultTransFlags = Flag_Trans_Sync_Mapping_Account

func (uc *UserConfig) GetFlagStatus(flag interface{}) bool {
	switch f := flag.(type) {
	case TransFlag:
		return uc.TransFlags&f > 0
	default:
		panic("Unknown [UserConfigFlag] type")
	}
}

func (uc *UserConfig) ForUpdate(tx *gorm.DB) error {
	if err := tx.Model(uc).Clauses(clause.Locking{Strength: "UPDATE"}).First(uc, uc.ID).Error; err != nil {
		return err
	}
	return nil
}

func (uc *UserConfig) ForShare(tx *gorm.DB) error {
	if err := tx.Model(uc).Clauses(clause.Locking{Strength: "SHARE"}).First(uc, uc.ID).Error; err != nil {
		return err
	}
	return nil
}

func (uc *UserConfig) OpenUserConfigFlag(flag interface{}, tx *gorm.DB) error {
	if true == uc.GetFlagStatus(flag) {
		return nil
	}
	var fileName = getUserConfigFlagName(flag)
	err := tx.Model(uc).Clauses(clause.Returning{}).Update(fileName, gorm.Expr(fileName+" | ?", flag)).Error
	if err != nil {
		return err
	}
	return tx.Model(uc).Select(fileName).First(uc).Error
}

func (uc *UserConfig) CloseUserConfigFlag(flag interface{}, tx *gorm.DB) error {
	if false == uc.GetFlagStatus(flag) {
		return nil
	}
	var fileName = getUserConfigFlagName(flag)
	err := tx.Model(uc).Update(fileName, gorm.Expr(fileName+" ^ ?", flag)).Error
	if err != nil {
		return err
	}
	return tx.Model(uc).Select(fileName).First(uc).Error
}

func getUserConfigFlagName(flag interface{}) string {
	switch flag.(type) {
	case TransFlag:
		return "trans_flags"
	default:
		panic("Unknown [UserConfigFlag] type")
	}
}
