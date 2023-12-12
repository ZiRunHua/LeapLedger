package global

type AmountCount struct {
	Amount int64
	Count  int64
}

type IncomeExpenseStatistic struct {
	Income  AmountCount
	Expense AmountCount
}
