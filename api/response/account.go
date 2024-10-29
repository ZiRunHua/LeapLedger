package response

import (
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/db"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	"github.com/ZiRunHua/LeapLedger/util/dataTool"
	"github.com/pkg/errors"
	"time"
)

type AccountOne struct {
	Id         uint
	Name       string
	Icon       string
	Type       accountModel.Type
	Location   string
	CreateTime time.Time
	UpdateTime time.Time
}

func (a *AccountOne) SetData(data accountModel.Account) error {
	a.Id = data.ID
	a.Name = data.Name
	a.Icon = data.Icon
	a.Type = data.Type
	a.Location = data.Location
	a.CreateTime = data.CreatedAt
	a.UpdateTime = data.UpdatedAt
	return nil
}

// AccountDetail 账本详情

type AccountDetail struct {
	AccountOne
	CreatorId   uint
	CreatorName string
	Role        accountModel.UserRole
	JoinTime    time.Time
}

func (a *AccountDetail) SetData(accountUser accountModel.User) error {
	// 账本
	account, err := accountUser.GetAccount()
	if err != nil {
		return err
	}
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
	a.JoinTime = accountUser.CreatedAt
	return nil
}

func (a *AccountDetail) SetDataFromAccountAndUser(account accountModel.Account, user userModel.User) error {
	a.setAccount(account)
	creator, err := account.GetUser("username")
	if err != nil {
		return err
	}
	a.CreatorName = creator.Username

	var accountUser accountModel.User
	accountUser, err = accountModel.NewDao().SelectUser(account.ID, user.ID)
	if err != nil {
		return err
	}
	a.Role = accountUser.GetRole()
	a.JoinTime = accountUser.CreatedAt
	return nil
}

// SetDataFromAccount 通过account设置数据，数据中的user来源为account.UserId
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
	a.JoinTime = accountUser.CreatedAt
	return nil
}

func (a *AccountDetail) setAccount(account accountModel.Account) {
	a.Id = account.ID
	a.Name = account.Name
	a.Icon = account.Icon
	a.Type = account.Type
	a.Location = account.Location
	a.CreatorId = account.UserId
	a.CreateTime = account.CreatedAt
	a.UpdateTime = account.UpdatedAt
}

type AccountDetailList []AccountDetail

func (a *AccountDetailList) SetData(list dataTool.Slice[uint, accountModel.User]) error {
	if len(list) == 0 {
		*a = make([]AccountDetail, 0)
		return nil
	}
	// 账本
	ids := list.ExtractValues(func(user accountModel.User) uint { return user.AccountId })
	var accountList dataTool.Slice[uint, accountModel.Account]
	err := db.Db.Where("id IN (?)", ids).Find(&accountList).Error
	if err != nil {
		return err
	}
	// 创建者
	ids = accountList.ExtractValues(func(account accountModel.Account) uint { return account.UserId })
	var creatorList dataTool.Slice[uint, userModel.User]
	err = db.Db.Select("username", "id").Where("id IN (?)", ids).Find(&creatorList).Error
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
		(*a)[i].JoinTime = user.CreatedAt
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
	RelatedAccount AccountDetail
	CreateTime     time.Time
	UpdateTime     time.Time
}

func (a *AccountMapping) SetData(data accountModel.Mapping) error {
	a.Id = data.ID
	a.CreateTime = data.CreatedAt
	a.UpdateTime = data.UpdatedAt
	account, err := data.GetMainAccount(db.Db)
	if err != nil {
		return err
	}
	err = a.MainAccount.SetData(account)
	if err != nil {
		return err
	}
	account, err = data.GetRelatedAccount(db.Db)
	if err != nil {
		return err
	}
	err = a.RelatedAccount.SetDataFromAccount(account)
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
	CreateTime time.Time
}

func (a *AccountUserInvitation) SetData(data accountModel.UserInvitation) error {
	var err error
	a.Id = data.ID
	a.Status = data.Status
	a.CreateTime = data.CreatedAt
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
	CreateTime time.Time
}

func (a *AccountUser) SetData(data accountModel.User) error {
	var err error
	a.Id = data.ID
	a.AccountId = data.AccountId
	a.UserId = data.UserId
	a.CreateTime = data.CreatedAt
	a.Role = data.GetRole()
	var info userModel.UserInfo
	if info, err = data.GetUserInfo(); err != nil {
		return err
	}
	a.Info.SetMaskData(info)
	return nil
}

type AccountUserInfo struct {
	TodayTransTotal        *global.IEStatisticWithTime
	CurrentMonthTransTotal *global.IEStatisticWithTime
	RecentTrans            *TransactionDetailList
}

type AccountInfo struct {
	TodayTransTotal        *global.IEStatisticWithTime
	CurrentMonthTransTotal *global.IEStatisticWithTime
	RecentTrans            *TransactionDetailList
}

type AccountUserConfig struct {
	Id        uint
	AccountId uint
	UserId    uint
	Trans     struct {
		SyncMappingAccount bool
	}
	CreateTime time.Time
	UpdateTime time.Time
}

func (auc *AccountUserConfig) SetData(data accountModel.UserConfig) error {
	auc.Id = data.ID
	auc.AccountId = data.AccountId
	auc.UserId = data.UserId
	auc.CreateTime = data.CreatedAt
	auc.UpdateTime = data.UpdatedAt
	auc.Trans.SyncMappingAccount = data.GetFlagStatus(accountModel.Flag_Trans_Sync_Mapping_Account)
	return nil
}
