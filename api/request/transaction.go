package request

import (
	"KeepAccount/global/constant"
)

type TransactionCreateOne struct {
	AccountId     uint                   `json:"account_id"`
	Amount        int                    `json:"amount"`
	CategoryId    uint                   `json:"category_id"`
	IncomeExpense constant.IncomeExpense `json:"income_expense"`
	Remark        string                 `json:"remark"`
	TradeTime     uint                   `json:"trade_time"`
}

type TransactionUpdateOne struct {
	AccountId     uint                   `json:"account_id"`
	Amount        int                    `json:"amount"`
	CategoryId    uint                   `json:"category_id"`
	IncomeExpense constant.IncomeExpense `json:"income_expense"`
	Remark        string                 `json:"remark"`
	TradeTime     uint                   `json:"trade_time"`
}

type TransactionGetList struct {
	IncomeExpense constant.IncomeExpense `json:"income_expense"`
	CategoryId    uint
	StartTime     uint
	EndTime       uint
	PageData      `json:"page_data"`
}
