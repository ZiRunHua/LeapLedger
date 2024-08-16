package response

import (
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"errors"
	"gorm.io/gorm"
	"time"
)

type Login struct {
	Token               string
	TokenExpirationTime time.Time
	CurrentAccount      AccountDetail
	CurrentShareAccount AccountDetail
	User                UserOne
}

func (l *Login) SetDataFormClientInto(data userModel.UserClientBaseInfo) error {
	var err error
	accountDao := accountModel.NewDao()
	// 当前客户端操作账本
	if data.CurrentAccountId != 0 {
		var account accountModel.Account
		account, err = accountDao.SelectById(data.CurrentAccountId)
		if err != nil {
			return err
		}
		err = l.CurrentAccount.SetDataFromAccount(account)
		if err != nil {
			return err
		}
	}
	// 当前客户端操作共享账本
	if data.CurrentShareAccountId != 0 {
		var accountUser accountModel.User
		accountUser, err = accountDao.SelectUser(data.CurrentShareAccountId, data.UserId)
		if err != nil && false == errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		err = l.CurrentShareAccount.SetData(accountUser)
		if err != nil {
			return err
		}
	}
	return nil
}

type Register struct {
	User                UserOne
	Token               string
	TokenExpirationTime time.Time
}

type UserOne struct {
	Id         uint
	Username   string
	Email      string
	CreateTime time.Time
}

func (u *UserOne) SetData(data userModel.User) error {
	u.Id = data.ID
	u.Email = data.Email
	u.Username = data.Username
	u.CreateTime = data.CreatedAt
	return nil
}

type UserHome struct {
	HeaderCard           *UserHomeHeaderCard
	TimePeriodStatistics *UserHomeTimePeriodStatistics
}

type UserHomeHeaderCard struct {
	*TransactionStatistic
}

type UserHomeTimePeriodStatistics struct {
	TodayData     *TransactionStatistic
	YesterdayData *TransactionStatistic
	WeekData      *TransactionStatistic
	YearData      *TransactionStatistic
}

type UserTransactionShareConfig struct {
	Account    bool
	CreateTime bool
	UpdateTime bool
	Remark     bool
}

func (u *UserTransactionShareConfig) SetData(data userModel.TransactionShareConfig) error {
	u.Account = data.GetFlagStatus(userModel.FLAG_ACCOUNT)
	u.CreateTime = data.GetFlagStatus(userModel.FLAG_CREATE_TIME)
	u.UpdateTime = data.GetFlagStatus(userModel.FLAG_UPDATE_TIME)
	u.Remark = data.GetFlagStatus(userModel.FLAG_REMARK)
	return nil
}

type UserFriendInvitation struct {
	Id         uint
	Inviter    UserInfo
	Invitee    UserInfo
	CreateTime time.Time
}

type UserInfo struct {
	Id       uint
	Username string
	Email    string
}

func (u *UserInfo) SetMaskData(data userModel.UserInfo) {
	u.Id = data.ID
	u.Username = data.Username
	u.Email = util.Str.MaskEmail(data.Email)
}

type UserCurrentClientInfo struct {
	CurrentAccount      AccountDetail
	CurrentShareAccount AccountDetail
	LoginTime           time.Time
}

func (u *UserCurrentClientInfo) SetData(info userModel.UserClientBaseInfo) error {
	u.LoginTime = info.LoginTime
	var accountUser accountModel.User
	var err error
	if info.CurrentAccountId > 0 {
		accountUser, err = accountModel.NewDao().SelectUser(info.CurrentAccountId, info.UserId)
		if err != nil {
			return err
		}
		err = u.CurrentAccount.SetData(accountUser)
		if err != nil {
			return err
		}
	}
	if info.CurrentShareAccountId > 0 {
		accountUser, err = accountModel.NewDao().SelectUser(info.CurrentShareAccountId, info.UserId)
		if err != nil {
			return err
		}
		err = u.CurrentShareAccount.SetData(accountUser)
		if err != nil {
			return err
		}
	}
	return nil
}
