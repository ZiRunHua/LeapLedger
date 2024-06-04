package transactionModel

import (
	"KeepAccount/global"
)

func init() {
	tables := []interface{}{
		&Transaction{}, &Mapping{},
		&ExpenseAccountStatistic{}, &ExpenseAccountUserStatistic{}, &ExpenseCategoryStatistic{},
		&IncomeAccountStatistic{}, &IncomeAccountUserStatistic{}, &IncomeCategoryStatistic{},
	}
	err := global.GvaDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
