package test

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	_ "KeepAccount/global/task"
	_ "KeepAccount/initialize"
	_ "KeepAccount/initialize/database"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
	_service "KeepAccount/service"
)

var (
	db                  = global.GvaDb
	User                userModel.User
	Account             accountModel.Account
	ExpenseCategoryList []categoryModel.Category
	transService        = _service.GroupApp.TransactionServiceGroup
)

func init() {
	var err error
	User, err = userModel.NewDao().SelectById(global.TestUserId)
	if err != nil {
		panic(err)
	}
	userInfo, err := User.GetUserClient(constant.Web)
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
