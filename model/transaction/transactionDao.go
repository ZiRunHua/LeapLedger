package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type TransactionDao struct {
	db *gorm.DB
}

func NewDao(db ...*gorm.DB) *TransactionDao {
	if len(db) > 0 {
		return &TransactionDao{db: db[0]}
	}
	return &TransactionDao{global.GvaDb}
}

func (t *TransactionDao) SelectById(id uint, forUpdate bool) (result Transaction, err error) {
	if forUpdate {
		err = t.db.Set("gorm:query_option", "FOR UPDATE").First(&result, id).Error
	} else {
		err = t.db.First(&result, id).Error
	}
	return
}

func (t *TransactionDao) GetListByCondition(condition Condition, limit int, offset int) (
	result []Transaction, err error,
) {
	query := condition.addConditionToQuery(t.db)
	err = query.Limit(limit).Offset(offset).Order("trade_time DESC").Find(&result).Error
	return
}

func (t *TransactionDao) GetIeStatisticByCondition(
	ie *constant.IncomeExpense, condition StatisticCondition, extCond *ExtensionCondition,
) (result global.IncomeExpenseStatistic, err error) {
	if extCond.IsConditionSet() {
		// 走transaction表查询
		query := condition.addConditionToQuery(t.db)
		query = extCond.addConditionToQuery(query)
		result, err = t.getIncomeExpenseStatisticByWhere(ie, query)
	} else {
		// 走统计表查询
		result, err = NewStatisticDao(t.db).GetIeStatisticByCondition(ie, condition)
	}
	if err != nil {
		err = errors.Wrap(err, "transactionDao.GetIeStatisticByCondition")
	}
	return
}

func (t *TransactionDao) getIncomeExpenseStatisticByWhere(ie *constant.IncomeExpense, query *gorm.DB) (
	result global.IncomeExpenseStatistic, err error,
) {
	if ie.QueryIncome() {
		//查询收入
		result.Income, err = t.getAmountCountStatistic(query, constant.Income)
		if err != nil {
			return
		}
	}
	if ie.QueryExpense() {
		//查询支出
		result.Expense, err = t.getAmountCountStatistic(query, constant.Expense)
		if err != nil {
			return
		}
	}
	return
}

func (t *TransactionDao) getAmountCountStatistic(query *gorm.DB, ie constant.IncomeExpense) (
	result global.AmountCount, err error,
) {
	err = query.Where("income_expense = ? ", ie).Select("COUNT(*) as Count,SUM(amount) as Amount").Scan(&result).Error
	return
}
