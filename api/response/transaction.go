package response

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	transactionModel "KeepAccount/model/transaction"
)

func TransactionModelToResponse(trans *transactionModel.Transaction) *TransactionOne {
	return &TransactionOne{
		Id:            trans.ID,
		UserId:        trans.UserId,
		AccountId:     trans.AccountId,
		Amount:        trans.Amount,
		CategoryId:    trans.CategoryId,
		IncomeExpense: trans.IncomeExpense,
		Remark:        trans.Remark,
		TradeTime:     trans.TradeTime.Unix(),
		UpdateTime:    trans.UpdatedAt.Unix(),
		CreateTime:    trans.CreatedAt.Unix(),
	}
}

type TransactionOne struct {
	Id            uint
	UserId        uint
	AccountId     uint
	Amount        int
	CategoryId    uint
	IncomeExpense constant.IncomeExpense
	Remark        string
	TradeTime     int64
	UpdateTime    int64
	CreateTime    int64
}

type TransactionDetail struct {
	Id                 uint
	UserId             uint
	UserName           string
	AccountId          uint
	AccountName        string
	Amount             int
	CategoryId         uint
	CategoryIcon       string
	CategoryName       string
	CategoryFatherName string
	IncomeExpense      constant.IncomeExpense
	Remark             string
	TradeTime          int64
	UpdateTime         int64
	CreateTime         int64
}

type TransactionGetList struct {
	List []TransactionDetail
	PageData
}
type TransactionTotal struct {
	global.IncomeExpenseStatistic
}

type TransactionStatistic struct {
	global.IncomeExpenseStatistic
	StartTime int64
	EndTime   int64
}

type TransactionMonthStatistic struct {
	List []TransactionStatistic
}

type TransactionDayStatistic struct {
	global.AmountCount
	Date int64
}

type TransactionCategoryAmountRank struct {
	Category CategoryOne
	global.AmountCount
}
