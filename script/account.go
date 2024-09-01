package script

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/cusCtx"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util/dataTool"
	"context"
	"encoding/json"
	"gorm.io/gorm"
	"io"
	"os"
)

type accountScripts struct {
}

var Account = accountScripts{}

func (as *accountScripts) CreateByTemplate(tmpl AccountTmpl, user userModel.User, ctx context.Context) (account accountModel.Account, accountUser accountModel.User, err error) {
	account, accountUser, err = accountService.CreateOne(user, tmpl.Name, tmpl.Icon, tmpl.Type, ctx)
	if err != nil {
		return
	}
	var list dataTool.Slice[any, fatherTmpl] = tmpl.Category
	for _, f := range list.CopyReverse() {
		err = f.create(account, ctx)
		if err != nil {
			return
		}
	}
	return
}

func (as *accountScripts) CreateExample(user userModel.User, ctx context.Context) (account accountModel.Account, accountUser accountModel.User, err error) {
	var accountTmpl AccountTmpl
	err = accountTmpl.ReadFromJson(constant.ExampleAccountJsonPath)
	if err != nil {
		return
	}
	return as.CreateByTemplate(accountTmpl, user, ctx)
}

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
	return accountService.CreateOne(user, at.Name, at.Icon, at.Type, cusCtx.WithDb(context.TODO(), tx))
}
