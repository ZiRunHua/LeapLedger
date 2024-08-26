package initialize

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	_ "KeepAccount/global/nats"
	_ "KeepAccount/initialize"
	_ "KeepAccount/initialize/database"
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

func init() {
	var err error
	User, err = userModel.NewDao().SelectById(global.TestUserId)
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
