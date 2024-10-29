package main

import (
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	commonModel "github.com/ZiRunHua/LeapLedger/model/common"

	_ "github.com/ZiRunHua/LeapLedger/initialize"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	"gorm.io/gorm"
)

func main() {
	correctStatistic()
}
func correctStatistic() {
	tables := []commonModel.Model{
		&transactionModel.ExpenseCategoryStatistic{}, &transactionModel.ExpenseAccountStatistic{},
		&transactionModel.ExpenseAccountUserStatistic{}, &transactionModel.IncomeCategoryStatistic{},
		&transactionModel.IncomeAccountStatistic{},
		&transactionModel.IncomeAccountUserStatistic{},
	}
	for _, table := range tables {
		global.GvaDb.Delete(table, "count >= ?", 0)
	}
	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return transAccumulate(tx)
		},
	)
	if err != nil {
		panic(err.Error())
	}
}

func transAccumulate(tx *gorm.DB) error {
	var list []transactionModel.Transaction
	err := tx.Model(&transactionModel.Transaction{}).Find(&list).Error
	if err != nil {
		return err
	}

	for _, trans := range list {
		if trans.IncomeExpense == constant.Expense {
			err = transactionModel.ExpenseAccumulate(
				trans.TradeTime, trans.AccountId, trans.UserId, trans.CategoryId, trans.Amount, 1, tx,
			)
			if err != nil {
				return err
			}
		} else {
			err = transactionModel.IncomeAccumulate(
				trans.TradeTime, trans.AccountId, trans.UserId, trans.CategoryId, trans.Amount, 1, tx,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
