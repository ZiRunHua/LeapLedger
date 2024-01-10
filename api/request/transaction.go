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
	UserId        uint
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

type TransactionDayStatistic struct {
	AccountId     uint `binding:"required"`
	CategoryIds   *[]uint
	IncomeExpense *constant.IncomeExpense `binding:"omitempty,oneof=income expense"`
	TimeFrame
}

type TransactionCategoryAmountRank struct {
	AccountId     uint                   `binding:"required"`
	IncomeExpense constant.IncomeExpense `binding:"required,oneof=income expense"`
	Limit         int                    `binding:"required"`
	TimeFrame
}
