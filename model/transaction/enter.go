package transactionModel

import (
	"KeepAccount/global"
)

type dao struct {
}

var (
	Dao = &dao{}
)

func init() {
	tables := []interface{}{
		IncomeCategoryStatistic{}, IncomeAccountStatistic{}, IncomeAccountUserStatistic{}, ExpenseCategoryStatistic{},
		ExpenseAccountStatistic{}, ExpenseAccountUserStatistic{},
	}
	for _, table := range tables {
		err := global.GvaDb.AutoMigrate(&table)
		if err != nil {
			panic(err)
		}
	}
}
