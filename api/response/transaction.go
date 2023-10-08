package response

import (
	"KeepAccount/global/constant"
	transactionModel "KeepAccount/model/transaction"
)

func TransactionModelToResponse(trans *transactionModel.Transaction) *TransactionOne {
	return &TransactionOne{
		Id:            trans.ID,
		AccountId:     trans.AccountID,
		Amount:        trans.Amount,
		CategoryId:    trans.CategoryID,
		IncomeExpense: trans.IncomeExpense,
		Remark:        trans.Remark,
		TradeTime:     trans.TradeTime.Unix(),
		UpdateTime:    trans.UpdatedAt.Unix(),
		CreateTime:    trans.CreatedAt.Unix(),
	}
}

type TransactionOne struct {
	Id            uint
	AccountId     uint
	Amount        int
	CategoryId    uint
	IncomeExpense constant.IncomeExpense
	Remark        string
	TradeTime     int64
	UpdateTime    int64
	CreateTime    int64
}

type TransactionGetList struct {
	List []TransactionOne
	PageData
}
