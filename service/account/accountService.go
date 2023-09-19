package accountService

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"github.com/pkg/errors"
)

type base struct{}

func (b *base) CreateOne(user *userModel.User, name string) (*accountModel.Account, error) {
	if name == "" {
		name = "账本"
	}
	account := &accountModel.Account{
		UserId: user.ID,
		Name:   name,
	}
	err := global.GvaDb.Create(account).Error
	return account, errors.Wrap(err, "Create(account)")
}

func (b *base) Delete(account *accountModel.Account) error {
	return account.GetDb().Delete(account).Error
}

func (b *base) Update(account *accountModel.Account, name string) error {
	if name == "" {
		return errors.Wrap(global.ErrInvalidParameter, "name")
	}
	err := global.GvaDb.Model(account).Update("name", name).Error
	if err != nil {
		return errors.Wrap(err, "account update")
	}
	return nil
}
