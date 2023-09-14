package transactionModel

import (
	"KeepAccount/global"
	"gorm.io/gorm"
)

type ExpenseStatisticDao struct {
	db *gorm.DB
}

func (d *dao) NewExpenseStatisticDao(db *gorm.DB) *ExpenseStatisticDao {
	if db == nil {
		db = global.GvaDb
	}
	return &ExpenseStatisticDao{db}
}
