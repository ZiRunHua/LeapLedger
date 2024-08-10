package test

import (
	categoryModel "KeepAccount/model/category"
	"KeepAccount/test/initialize"
)

var (
	User                = initialize.User
	Account             = initialize.Account
	ExpenseCategoryList []categoryModel.Category
)

func init() {
	ExpenseCategoryList = initialize.ExpenseCategoryList
}
