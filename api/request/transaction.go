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
	AccountId     uint `binding:"required"`
	CategoryId    *uint
	IncomeExpense *constant.IncomeExpense `binding:"omitempty,oneof=income expense"`
	StartTime     *int64                  `binding:"omitempty,gt=0"`
	EndTime       *int64                  `binding:"omitempty,gt=0"`
	PageData
}
