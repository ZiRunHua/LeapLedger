package accountService

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type base struct{}

func (b *base) CreateOne(user *userModel.User, name string, icon string, tx *gorm.DB) (*accountModel.Account, error) {
	if name == "" || icon == "" {
		return nil, global.NewErrDataIsEmpty("name或icon")
	}
	account := &accountModel.Account{
		UserId: user.ID,
		Name:   name,
		Icon:   icon,
	}
	account.SetTx(tx)
	err := account.CreateOne()
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
