package request

import (
	"KeepAccount/global/constant"
)

type TransactionCreateOne struct {
	AccountId     uint
	Amount        int
	CategoryId    uint
	IncomeExpense constant.IncomeExpense
	Remark        string
	TradeTime     uint
}

type TransactionUpdateOne struct {
	AccountId     uint
	Amount        int
	CategoryId    uint
	IncomeExpense constant.IncomeExpense
	Remark        string
	TradeTime     uint
}

type TransactionGetList struct {
	UserId        *uint
	AccountId     *uint
	CategoryId    *uint
	IncomeExpense *constant.IncomeExpense
	StartTime     *int64
	EndTime       *int64
	PageData
}
