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

type TransactionQueryCondition struct {
	AccountId     uint `binding:"required"`
	UserIds       *[]uint
	CategoryIds   *[]uint
	IncomeExpense *constant.IncomeExpense `binding:"omitempty,oneof=income expense"`
	MinimumAmount *int                    `binding:"omitempty,min=0"`
	MaximumAmount *int                    `binding:"omitempty,min=0"`
	TimeFrame
}

type TransactionGetList struct {
	TransactionQueryCondition
	PageData
}

type TransactionTotal struct {
	TransactionQueryCondition
}

type TransactionMonthStatistic struct {
	TransactionQueryCondition
}
