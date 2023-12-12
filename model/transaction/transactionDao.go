package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type TransactionDao struct {
	db *gorm.DB
}
type _transactionDao interface {
	GetListByCondition(condition *TransactionCondition, limit int, offset int) (result []Transaction, err error)
	GetStatisticByCondition(
		condition *StatisticCondition,
		TradeStartTime time.Time,
		TradeEndTime time.Time,
	) (result *global.IncomeExpenseStatistic, err error)
}

func (d *dao) NewTransaction(db *gorm.DB) *TransactionDao {
	if db == nil {
		db = global.GvaDb
	}
	return &TransactionDao{db}
}

type TradeTimeCondition struct {
	TradeStartTime *time.Time
	TradeEndTime   *time.Time
}

type ForeignKeyCondition struct {
	AccountId   *uint
	UserIds     *[]uint
	CategoryIds *[]uint
}

type TransactionCondition struct {
	ForeignKeyCondition
	TradeTimeCondition
	IncomeExpense *constant.IncomeExpense
	MinimumAmount *int
	MaximumAmount *int
}

func (t *TransactionDao) setQueryByForeignKeyCondition(condition *ForeignKeyCondition) {
	query := t.db
	if condition.AccountId != nil {
		query = query.Where("account_id = ?", *condition.AccountId)
	}
	if condition.UserIds != nil && len(*condition.UserIds) > 0 {
		query = query.Where("user_id IN (?)", *condition.UserIds)
	}
	if condition.CategoryIds != nil && len(*condition.CategoryIds) > 0 {
		query = query.Where("category_id IN (?)", *condition.CategoryIds)
	}
	t.db = query.Session(&gorm.Session{})
}

func (t *TransactionDao) setQueryByCondition(condition *TransactionCondition) {
	t.setQueryByForeignKeyCondition(&condition.ForeignKeyCondition)
	query := t.db
	if condition.IncomeExpense != nil {
		query = query.Where("income_expense = ?", *condition.IncomeExpense)
	}
	if condition.MinimumAmount != nil {
		query = query.Where("amount >= ?", *condition.MinimumAmount)
	}
	if condition.MaximumAmount != nil {
		query = query.Where("amount <= ?", *condition.MaximumAmount)
	}
	if condition.TradeStartTime != nil && condition.TradeEndTime != nil {
		query = query.Where("trade_time BETWEEN ? AND ?", *condition.TradeStartTime, *condition.TradeEndTime)
	}
	t.db = query.Session(&gorm.Session{})
}
func (t *TransactionDao) GetListByCondition(
	condition *TransactionCondition, limit int, offset int,
) (result []Transaction, err error) {
	t.setQueryByCondition(condition)
	err = t.db.Limit(limit).Offset(offset).Order("trade_time DESC").Find(&result).Error
	return
}

type StatisticCondition struct {
	ForeignKeyCondition
	IncomeExpense *constant.IncomeExpense
	MinimumAmount *int
	MaximumAmount *int
}

func (t *TransactionDao) GetStatisticByCondition(
	condition *StatisticCondition,
	TradeStartTime time.Time,
	TradeEndTime time.Time,
) (result *global.IncomeExpenseStatistic, err error) {
	t.setQueryByForeignKeyCondition(&condition.ForeignKeyCondition)
	query := t.db
	if condition.MinimumAmount != nil {
		query = query.Where("amount >= ?", *condition.MinimumAmount)
	}
	if condition.MaximumAmount != nil {
		query = query.Where("amount <= ?", *condition.MaximumAmount)
	}
	t.db = query.Session(&gorm.Session{})
	result, err = t.getIncomeExpenseStatisticByWhere(condition.IncomeExpense, TradeStartTime, TradeEndTime)
	if err != nil {
		err = errors.Wrap(err, "transactionDao.GetStatisticByCondition")
	}
	return
}

func (t *TransactionDao) getIncomeExpenseStatisticByWhere(
	incomeExpense *constant.IncomeExpense,
	TradeStartTime time.Time,
	TradeEndTime time.Time,
) (result *global.IncomeExpenseStatistic, err error) {
	result = &global.IncomeExpenseStatistic{
		Income:  global.AmountCount{},
		Expense: global.AmountCount{},
	}
	//根据传入的IncomeExpense条件判断查询 以减少查询次数
	if incomeExpense == nil {
		//查询收入和支出
		err = t.doAmountCountStatistic(incomeExpense, TradeStartTime, TradeEndTime, &result.Income)
		if err != nil {
			return
		}
		err = t.doAmountCountStatistic(incomeExpense, TradeStartTime, TradeEndTime, &result.Expense)
		if err != nil {
			return
		}
	} else if *incomeExpense == constant.Income {
		//查询收入
		err = t.doAmountCountStatistic(incomeExpense, TradeStartTime, TradeEndTime, &result.Income)
		if err != nil {
			return
		}
	} else {
		//查询支出
		err = t.doAmountCountStatistic(incomeExpense, TradeStartTime, TradeEndTime, &result.Expense)
		if err != nil {
			return
		}
	}
	return
}

func (t *TransactionDao) doAmountCountStatistic(
	incomeExpense *constant.IncomeExpense,
	TradeStartTime time.Time,
	TradeEndTime time.Time,
	result *global.AmountCount,
) (err error) {
	query := t.db.Where("income_expense = ?", incomeExpense)
	query = t.db.Where("trade_time BETWEEN ? AND ?", TradeStartTime, TradeEndTime)

	err = query.Model(&Transaction{}).Count(&result.Count).Error
	if err != nil {
		return
	}
	err = query.Model(&Transaction{}).Select("SUM(amount) as Amount").Scan(&result).Error
	if err != nil {
		return
	}
	return
}
