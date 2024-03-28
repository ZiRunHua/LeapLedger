package response

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util/dataType"
	"github.com/pkg/errors"
)

func AccountModelToResponse(account *accountModel.Account) AccountOne {
	if account == nil {
		return AccountOne{}
	}
	return AccountOne{
		Id:         account.ID,
		Name:       account.Name,
		Icon:       account.Icon,
		Type:       account.Type,
		UpdateTime: account.UpdatedAt.Unix(),
		CreateTime: account.CreatedAt.Unix(),
	}
}

type AccountOne struct {
	Id         uint
	Name       string
	Icon       string
	Type       accountModel.Type
	CreateTime int64
	UpdateTime int64
}

func (a *AccountOne) SetData(data accountModel.Account) error {
	a.Id = data.ID
	a.Name = data.Name
	a.Icon = data.Icon
	a.Type = data.Type
	a.CreateTime = data.CreatedAt.Unix()
	a.UpdateTime = data.UpdatedAt.Unix()
	return nil
}

// 账本详情
type AccountDetail struct {
	AccountOne
	CreatorId   uint
	CreatorName string
	Role        accountModel.UserRole
	JoinTime    int64
}

func (a *AccountDetail) SetData(accountUser accountModel.User) error {
	// 账本
	account, err := accountUser.GetAccount()
	if account.ID != accountUser.AccountId {
		return errors.New("accountUser not belong account")
	}
	a.setAccount(account)
	// 创建者
	var user userModel.User
	user, err = account.GetUser("username")
	if err != nil {
		return err
	}
	a.CreatorName = user.Username

	a.Role = accountUser.GetRole()
	a.JoinTime = accountUser.CreatedAt.Unix()
	return nil
}
func (a *AccountDetail) SetDataFromAccount(account accountModel.Account) error {
	a.setAccount(account)

	user, err := account.GetUser("username", "id")
	if err != nil {
		return err
	}
	a.CreatorName = user.Username

	var accountUser accountModel.User
	accountUser, err = accountModel.NewDao().SelectUser(account.ID, user.ID)
	if err != nil {
		return err
	}
	a.Role = accountUser.GetRole()
	a.JoinTime = accountUser.CreatedAt.Unix()
	return nil
}

func (a *AccountDetail) setAccount(account accountModel.Account) {
	a.Id = account.ID
	a.Name = account.Name
	a.Icon = account.Icon
	a.Type = account.Type
	a.CreatorId = account.UserId
	a.CreateTime = account.CreatedAt.Unix()
	a.UpdateTime = account.UpdatedAt.Unix()
}

type AccountDetailList []AccountDetail

func (a *AccountDetailList) SetData(list dataType.Slice[uint, accountModel.User]) error {
	if len(list) == 0 {
		*a = make([]AccountDetail, 0, 0)
		return nil
	}
	// 账本
	ids := list.ExtractValues(func(user accountModel.User) uint { return user.AccountId })
	var accountList dataType.Slice[uint, accountModel.Account]
	err := global.GvaDb.Where("id IN (?)", ids).Find(&accountList).Error
	if err != nil {
		return err
	}
	// 创建者
	ids = accountList.ExtractValues(func(account accountModel.Account) uint { return account.UserId })
	var creatorList dataType.Slice[uint, userModel.User]
	err = global.GvaDb.Select("username", "id").Where("id IN (?)", ids).Find(&creatorList).Error
	if err != nil {
		return err
	}

	userMap := list.ToMap(func(user accountModel.User) uint { return user.AccountId })
	creatorMap := creatorList.ToMap(func(user userModel.User) uint { return user.ID })
	*a = make([]AccountDetail, len(accountList), len(accountList))
	for i, account := range accountList {
		(*a)[i].setAccount(account)
		(*a)[i].CreatorName = creatorMap[account.UserId].Username
		user := userMap[account.ID]
		(*a)[i].Role = user.GetRole()
		(*a)[i].JoinTime = user.CreatedAt.Unix()
	}
	return nil
}

type AccountTemplateOne struct {
	Id   uint
	Name string
	Icon string
	Type accountModel.Type
}

type AccountTemplateList struct {
	List []AccountTemplateOne
}

// AccountMapping 账本关联
type AccountMapping struct {
	Id             uint
	MainAccount    AccountOne
	RelatedAccount AccountOne
	CreateTime     int64
	UpdateTime     int64
}

func (a *AccountMapping) SetData(data accountModel.Mapping) error {
	a.Id = data.ID
	a.CreateTime = data.CreatedAt.Unix()
	a.UpdateTime = data.UpdatedAt.Unix()
	account, err := data.GetMainAccount()
	if err != nil {
		return err
	}
	err = a.MainAccount.SetData(account)
	if err != nil {
		return err
	}
	account, err = data.GetRelatedAccount()
	if err != nil {
		return err
	}
	err = a.RelatedAccount.SetData(account)
	if err != nil {
		return err
	}
	return nil
}

type AccountUserInvitation struct {
	Id         uint
	Account    AccountOne
	Inviter    UserInfo
	Invitee    UserInfo
	Status     accountModel.UserInvitationStatus
	Role       accountModel.UserRole
	CreateTime int64
}

func (a *AccountUserInvitation) SetData(data accountModel.UserInvitation) error {
	var err error
	a.Id = data.ID
	a.Status = data.Status
	a.CreateTime = data.CreatedAt.Unix()
	a.Role = data.GetRole()

	var account accountModel.Account
	if account, err = data.GetAccount(); err != nil {
		return err
	}
	err = a.Account.SetData(account)
	if err != nil {
		return err
	}
	var info userModel.UserInfo
	if info, err = data.GetInviterInfo(); err != nil {
		return err
	}
	a.Inviter.SetMaskData(info)
	if info, err = data.GetInviteeInfo(); err != nil {
		return err
	}
	a.Invitee.SetMaskData(info)
	return nil
}

type AccountUser struct {
	Id         uint
	AccountId  uint
	UserId     uint
	Info       UserInfo
	Role       accountModel.UserRole
	CreateTime int64
}

func (a *AccountUser) SetData(data accountModel.User) error {
	var err error
	a.Id = data.ID
	a.AccountId = data.AccountId
	a.UserId = data.UserId
	a.CreateTime = data.CreatedAt.Unix()
	a.Role = data.GetRole()
	var info userModel.UserInfo
	if info, err = data.GetUserInfo(); err != nil {
		return err
	}
	a.Info.SetMaskData(info)
	return nil
}

type AccountUserInfo struct {
	TodayTransTotal        *global.IncomeExpenseStatistic
	CurrentMonthTransTotal *global.IncomeExpenseStatistic
	RecentTrans            *TransactionDetailList
}

type AccountInfo struct {
	TodayTransTotal        *global.IncomeExpenseStatistic
	CurrentMonthTransTotal *global.IncomeExpenseStatistic
	RecentTrans            *TransactionDetailList
}
