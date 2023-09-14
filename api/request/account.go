package request

import accountModel "KeepAccount/model/account"

// AccountCreateOne 账本新建
type AccountCreateOne struct {
	Name string            `binding:"required"`
	Icon string            `binding:"required"`
	Type accountModel.Type `binding:"required"`
}

// AccountUpdateOne 账本修改
type AccountUpdateOne struct {
	Name *string
	Icon *string
	Type accountModel.Type `binding:"required"`
}

// AccountTransCategoryInit 账本交易类型初始话化
type AccountTransCategoryInit struct {
	TemplateId uint
}

// AccountMapping 账本关联
type AccountMapping struct {
	AccountId uint
}

// UpdateAccountMapping 账本关联
type UpdateAccountMapping struct {
	RelatedAccountId uint
}

// AccountCreateOneUserInvitation 账本邀请建立
type AccountCreateOneUserInvitation struct {
	Invitee uint                   `binding:"required"`
	Role    *accountModel.UserRole `binding:"omitempty"`
}

// AccountGetUserInvitationList 账本邀请列表
type AccountGetUserInvitationList struct {
	AccountId uint                   `binding:"required"`
	Invitee   *uint                  `binding:"omitempty"`
	Role      *accountModel.UserRole `binding:"omitempty"`
	PageData
}

// AccountGetUserInfo 账本用户信息获取
type AccountGetUserInfo struct {
	Types []InfoType
}

type AccountInfo struct {
	Types *[]InfoType `binding:"omitempty"`
}

type AccountUpdateUser struct {
	Role accountModel.UserRole `binding:"required"`
}

func (a *AccountUpdateUser) GetUpdateData() accountModel.UserUpdateData {
	return accountModel.UserUpdateData{
		Permission: a.Role.ToUserPermission(),
	}
}
