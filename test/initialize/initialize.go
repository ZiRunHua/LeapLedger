package initialize

import (
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/global/db"
	_ "github.com/ZiRunHua/LeapLedger/global/nats"
	"github.com/ZiRunHua/LeapLedger/global/nats/manager"
	_ "github.com/ZiRunHua/LeapLedger/initialize/database"
	"github.com/ZiRunHua/LeapLedger/test/info"
)
import (
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	categoryModel "github.com/ZiRunHua/LeapLedger/model/category"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
)

var (
	User                userModel.User
	Account             accountModel.Account
	ExpenseCategoryList []categoryModel.Category
)

var (
	Info = info.Data
)

func init() {
	manager.UpdateTestBackOff()
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
