package global

import "KeepAccount/util/timeTool"

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
	StartTime timeTool.Timestamp
	EndTime   timeTool.Timestamp
}
