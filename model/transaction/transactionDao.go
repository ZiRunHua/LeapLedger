package transaction

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"gorm.io/gorm"
	"time"
)

type TransactionDao struct {
	db *gorm.DB
}

func NewTransactionDao(db *gorm.DB) *TransactionDao {
	if db == nil {
		db = global.GvaDb
	}
	return &TransactionDao{db}
}

type _transactionDao interface {
	GetListByCondition(condition *TransactionCondition, limit int, offset int) (result *[]Transaction, err error)
}
type TransactionCondition struct {
	UserID         *uint
	AccountID      *uint
	CategoryID     *uint
	IncomeExpense  *constant.IncomeExpense
	TradeStartTime *time.Time
	TradeEndTime   *time.Time
}

func (t *TransactionDao) GetListByCondition(
	condition *TransactionCondition, limit int, offset int,
) (result *[]Transaction, err error) {
	query := t.db
	where := &Transaction{}
	if condition.UserID != nil {
		where.UserID = *condition.UserID
	}
	if condition.AccountID != nil {
		where.AccountID = *condition.AccountID
	}
	if condition.CategoryID != nil {
		where.CategoryID = *condition.CategoryID
	}
	if condition.IncomeExpense != nil {
		where.IncomeExpense = *condition.IncomeExpense
	}
	query = query.Where(where)
	if condition.TradeStartTime != nil && condition.TradeEndTime != nil {
		query = query.Where("trade_time BETWEEN ? AND ?", *condition.TradeStartTime, *condition.TradeEndTime)
	}
	err = query.Find(&result).Limit(limit).Offset(offset).Error
	return
}
