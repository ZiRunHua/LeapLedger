package test

import (
	categoryModel "github.com/ZiRunHua/LeapLedger/model/category"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	"github.com/ZiRunHua/LeapLedger/util/rand"
	"time"
)

func getCategory() categoryModel.Category {
	return ExpenseCategoryList[0]
}
func NewTransaction() transactionModel.Transaction {
	return transactionModel.Transaction{
		Info: NewTransInfo(),
	}
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
