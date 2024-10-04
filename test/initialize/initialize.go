package initialize

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	_ "KeepAccount/global/nats"
	_ "KeepAccount/initialize/database"
	"KeepAccount/test/info"
)
import (
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
)

var (
	User                userModel.User
	Account             accountModel.Account
	ExpenseCategoryList []categoryModel.Category
)

const (
	port = "7979"
	Host = "127.0.0.1:" + port
)

var (
	Info = info.Data
)

func init() {
	var err error
	User, err = userModel.NewDao().SelectById(Info.UserId)
	if err != nil {
		panic(err)
	}
	userInfo, err := User.GetUserClient(constant.Web, db.Db)
	if err != nil {
		panic(err)
	}
	Account, err = accountModel.NewDao().SelectById(userInfo.CurrentAccountId)
	if err != nil {
		panic(err)
	}
	ie := constant.Expense
	ExpenseCategoryList, err = categoryModel.NewDao().GetListByAccount(Account, &ie)
	if err != nil {
		panic(err)
	}
}
