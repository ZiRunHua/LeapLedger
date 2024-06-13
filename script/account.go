package script

import (
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"encoding/json"
	"gorm.io/gorm"
	"io"
	"os"
)

type _account struct {
}

var Account = _account{}

type AccountTmpl struct {
	Name, Icon string
	Type       accountModel.Type
	Category   []fatherTmpl
}

func (at *AccountTmpl) ReadFromJson(path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, at)
	if err != nil {
		return err
	}
	return nil
}

func (at *AccountTmpl) create(user userModel.User, tx *gorm.DB) (accountModel.Account, accountModel.User, error) {
	return accountService.CreateOne(user, at.Name, at.Icon, at.Type, tx)
}

func (u *_account) CreateByTemplate(tmpl AccountTmpl, user userModel.User, tx *gorm.DB) (account accountModel.Account, accountUser accountModel.User, err error) {
	account, accountUser, err = accountService.CreateOne(user, tmpl.Name, tmpl.Icon, tmpl.Type, tx)
	if err != nil {
		return
	}
	for _, f := range tmpl.Category {
		err = f.create(account, tx)
		if err != nil {
			return
		}
	}
	return
}

func (a *_account) CreateExample(user userModel.User, tx *gorm.DB) (account accountModel.Account, accountUser accountModel.User, err error) {
	var accountTmpl AccountTmpl
	err = accountTmpl.ReadFromJson(constant.ExampleAccountJsonPath)
	if err != nil {
		return
	}
	return a.CreateByTemplate(accountTmpl, user, tx)
}
