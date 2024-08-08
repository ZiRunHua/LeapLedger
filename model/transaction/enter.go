package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/util/gormFunc"
)

func CurrentInit() error {
	tables := []interface{}{
		Transaction{}, Mapping{},
		ExpenseAccountStatistic{}, ExpenseAccountUserStatistic{}, ExpenseCategoryStatistic{},
		IncomeAccountStatistic{}, IncomeAccountUserStatistic{}, IncomeCategoryStatistic{},
		Timing{}, TimingExec{},
	}
	err := global.GvaDb.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		_ = gormFunc.AlterIdToHeader(table, global.GvaDb)
	}
	return nil
}
