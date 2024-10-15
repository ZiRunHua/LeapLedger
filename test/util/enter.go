package tUtil

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
	_service "KeepAccount/service"
	"KeepAccount/test/initialize"
)

var (
	userService     = _service.GroupApp.UserServiceGroup
	templateService = _service.GroupApp.TemplateServiceGroup

	testUser         userModel.User
	testAccount      accountModel.Account
	testCategoryList []categoryModel.Category
	testInfo         = initialize.Info
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
	testCategoryList, err = categoryModel.NewDao().GetListByAccount(testAccount, nil)
	if err != nil {
		panic(err)
	}
}
