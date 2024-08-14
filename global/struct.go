package global

import "time"

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
	StartTime time.Time
	EndTime   time.Time
}
