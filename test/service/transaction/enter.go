package transaction

import (
	"KeepAccount/global/db"
	"KeepAccount/test/initialize"
	_ "KeepAccount/test/initialize"
)
import (
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
	_service "KeepAccount/service"
)

var (
	testUser                userModel.User
	testAccount             accountModel.Account
	testExpenseCategoryList []categoryModel.Category
	service                 = _service.GroupApp.TransactionServiceGroup
	testInfo                = initialize.Info
)

func init() {
	var err error
	testUser, err = userModel.NewDao().SelectById(testInfo.UserId)
	if err != nil {
		panic(err)
	}
	userInfo, err := testUser.GetUserClient(constant.Web, db.Db)
	if err != nil {
		panic(err)
	}
	testAccount, err = accountModel.NewDao().SelectById(userInfo.CurrentAccountId)
	if err != nil {
		panic(err)
	}
	ie := constant.Expense
	testExpenseCategoryList, err = categoryModel.NewDao().GetListByAccount(testAccount, &ie)
	if err != nil {
		panic(err)
	}
}

func getCategory() categoryModel.Category {
	return testExpenseCategoryList[0]
}
