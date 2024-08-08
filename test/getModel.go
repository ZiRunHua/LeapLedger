package test

import (
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/util/rand"
	"time"
)

func getCategory() categoryModel.Category {
	return ExpenseCategoryList[0]
}

func NewTransTime() transactionModel.Timing {
	transInfo := NewTransInfo()
	return transactionModel.Timing{
		AccountId:  transInfo.AccountId,
		UserId:     transInfo.UserId,
		TransInfo:  transInfo,
		Type:       transactionModel.EveryDay,
		OffsetDays: 1,
		NextTime:   transInfo.TradeTime,
		Close:      false,
	}
}

func NewTransInfo() transactionModel.Info {
	category := getCategory()
	return transactionModel.Info{
		UserId:        User.ID,
		AccountId:     Account.ID,
		CategoryId:    category.ID,
		IncomeExpense: category.IncomeExpense,
		Amount:        rand.Int(1000),
		Remark:        "test",
		TradeTime:     time.Now(),
	}
}
