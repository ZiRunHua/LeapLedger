package transactionModel

import (
	"KeepAccount/util/gormFunc"
	"gorm.io/gorm"
)

func CurrentInit(db *gorm.DB) error {
	tables := []interface{}{
		Transaction{}, Mapping{},
		ExpenseAccountStatistic{}, ExpenseAccountUserStatistic{}, ExpenseCategoryStatistic{},
		IncomeAccountStatistic{}, IncomeAccountUserStatistic{}, IncomeCategoryStatistic{},
		Timing{}, TimingExec{},
	}
	err := db.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		_ = gormFunc.AlterIdToHeader(table, db)
	}
	return nil
}
