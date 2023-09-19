package response

import (
	"KeepAccount/global/constant"
)

type TransactionOne struct {
	AccountId     uint                   `json:"account_id"`
	Amount        int                    `json:"amount"`
	CategoryId    uint                   `json:"category_id"`
	IncomeExpense constant.IncomeExpense `json:"income_expense"`
	Remark        string                 `json:"remark"`
	TradeTime     int64                  `json:"trade_time"`
}

type TransactionGetList struct {
	List []TransactionOne `json:"list"`
	PageData
}
