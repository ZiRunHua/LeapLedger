package transactionModel

import "KeepAccount/global/db"

func init() {
	tables := []interface{}{
		Transaction{}, Mapping{},
		ExpenseAccountStatistic{}, ExpenseAccountUserStatistic{}, ExpenseCategoryStatistic{},
		IncomeAccountStatistic{}, IncomeAccountUserStatistic{}, IncomeCategoryStatistic{},
		Timing{}, TimingExec{},
	}
	err := db.InitDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
