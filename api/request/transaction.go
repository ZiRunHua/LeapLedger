package request

import (
	"KeepAccount/global/constant"
	transactionModel "KeepAccount/model/transaction"
	"time"
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

func (t *TransactionQueryCondition) GetCondition() transactionModel.Condition {
	startTime := t.TimeFrame.GetStartTime()
	endTime := t.TimeFrame.GetEndTime()
	return transactionModel.Condition{
		IncomeExpense:       t.IncomeExpense,
		TimeCondition:       transactionModel.TimeCondition{TradeStartTime: &startTime, TradeEndTime: &endTime},
		ForeignKeyCondition: t.GetForeignKeyCondition(),
		ExtensionCondition:  t.GetExtensionCondition(),
	}
}

func (t *TransactionQueryCondition) GetForeignKeyCondition() transactionModel.ForeignKeyCondition {
	return transactionModel.ForeignKeyCondition{
		AccountId:   t.AccountId,
		UserIds:     t.UserIds,
		CategoryIds: t.CategoryIds,
	}
}

func (t *TransactionQueryCondition) GetStatisticCondition() transactionModel.StatisticCondition {
	return transactionModel.StatisticCondition{
		ForeignKeyCondition: t.GetForeignKeyCondition(),
		StartTime:           t.GetStartTime(),
		EndTime:             t.GetEndTime(),
	}
}

func (t *TransactionQueryCondition) GetExtensionCondition() transactionModel.ExtensionCondition {
	return transactionModel.ExtensionCondition{
		MinAmount: t.MinimumAmount,
		MaxAmount: t.MaximumAmount,
	}
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
	Limit         *int                   `binding:"omitempty"`
	TimeFrame
}

type TransactionAmountRank struct {
	AccountId     uint                   `binding:"required"`
	IncomeExpense constant.IncomeExpense `binding:"required,oneof=income expense"`
	TimeFrame
}

type TransactionTimingConfig struct {
	UserId     uint
	Type       transactionModel.TimingType
	OffsetDays int
	NextTime   time.Time
}

type TransactionTiming struct {
	Trans  transactionModel.Info
	Config TransactionTimingConfig
}

func (tt TransactionTiming) GetTimingModel() transactionModel.Timing {
	return transactionModel.Timing{
		TransInfo:  tt.Trans,
		UserId:     tt.Config.UserId,
		Type:       tt.Config.Type,
		OffsetDays: tt.Config.OffsetDays,
		NextTime:   tt.Config.NextTime,
	}
}
