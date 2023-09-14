package response

import (
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
)

type Login struct {
	Token               string
	CurrentAccount      AccountDetail
	CurrentShareAccount AccountDetail
	User                UserOne
}

type Register struct {
	User  UserOne
	Token string
}

type UserOne struct {
	Id         uint
	Username   string
	Email      string
	CreateTime int64
}

func (u *UserOne) SetData(data userModel.User) error {
	u.Id = data.ID
	u.Email = data.Email
	u.Username = data.Username
	u.CreateTime = data.CreatedAt.Unix()
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
	CreateTime int64
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
	LoginTime           int64
}

func (u *UserCurrentClientInfo) SetData(info userModel.UserClientBaseInfo) error {
	u.LoginTime = info.LoginTime.Unix()
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
