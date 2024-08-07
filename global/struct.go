package global

type AmountCount struct {
	Amount int64
	Count  int64
}

type IEStatistic struct {
	Income  AmountCount
	Expense AmountCount
}

type IEStatisticWithTime struct {
	IEStatistic
	StartTime int64
	EndTime   int64
}
