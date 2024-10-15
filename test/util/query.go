package tUtil

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	categoryModel "KeepAccount/model/category"
)

type Query struct {
}

func (q *Query) Category(accountID uint, ie constant.IncomeExpense) (category categoryModel.Category, err error) {
	err = db.Db.Where("account_id  = ? AND income_expense = ?", accountID, ie).First(&category).Error
	return
}
