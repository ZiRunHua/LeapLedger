package transactionModel

import (
	"KeepAccount/global"
	"gorm.io/gorm"
)

type IncomeStatisticDao struct {
	db *gorm.DB
}

func (d *dao) NewIncomeStatisticDao(db *gorm.DB) *IncomeStatisticDao {
	if db == nil {
		db = global.GvaDb
	}
	return &IncomeStatisticDao{db}
}
