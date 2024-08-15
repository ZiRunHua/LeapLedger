package request

import accountModel "KeepAccount/model/account"

// AccountCreateOne 账本新建
// swagger:parameters AccountCreateOne
type AccountCreateOne struct {
	Name string            `binding:"required"`
	Icon string            `binding:"required"`
	Type accountModel.Type `binding:"required"`
}

// AccountUpdateOne 账本修改
// swagger:parameters AccountUpdateOne
type AccountUpdateOne struct {
	Name *string
	Icon *string
	Type accountModel.Type `binding:"required"`
}

// AccountTransCategoryInit 账本交易类型初始话化
// swagger:parameters AccountTransCategoryInit
type AccountTransCategoryInit struct {
	TemplateId uint
}

// AccountMapping 账本关联
// swagger:parameters AccountMapping
type AccountMapping struct {
	AccountId uint
}

// UpdateAccountMapping 账本关联
// swagger:parameters UpdateAccountMapping
type UpdateAccountMapping struct {
	RelatedAccountId uint
}

// AccountCreateOneUserInvitation 账本邀请建立
// swagger:parameters AccountCreateOneUserInvitation
type AccountCreateOneUserInvitation struct {
	Invitee uint                   `binding:"required"`
	Role    *accountModel.UserRole `binding:"omitempty"`
}

// AccountGetUserInvitationList 账本邀请列表
// swagger:parameters AccountGetUserInvitationList
type AccountGetUserInvitationList struct {
	AccountId uint                   `binding:"required"`
	Invitee   *uint                  `binding:"omitempty"`
	Role      *accountModel.UserRole `binding:"omitempty"`
	PageData
}

// AccountGetUserInfo 账本用户信息获取
// swagger:parameters AccountGetUserInfo
type AccountGetUserInfo struct {
	Types []InfoType
}

// swagger:parameters AccountInfo
type AccountInfo struct {
	Types *[]InfoType `binding:"omitempty"`
}

// swagger:parameters AccountUpdateUser
type AccountUpdateUser struct {
	Role accountModel.UserRole `binding:"required"`
}

func (a *AccountUpdateUser) GetUpdateData() accountModel.UserUpdateData {
	return accountModel.UserUpdateData{
		Permission: a.Role.ToUserPermission(),
	}
}

// swagger:parameters AccountUserConfigFlagUpdate
type AccountUserConfigFlagUpdate struct {
	AccountId uint
	Status    bool
}
