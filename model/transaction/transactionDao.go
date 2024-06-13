package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
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
	if extCond.IsSet() {
		// 走transaction表查询
		query := condition.ForeignKeyCondition.addConditionToQuery(t.db)
		query = query.Where("trans_time between ? AND ?", condition.StartTime, condition.EndTime)
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
		result.Income, err = t.getAmountCountStatistic(query, constant.Income)
		if err != nil {
			return
		}
	}
	if ie.QueryExpense() {
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

func (t *TransactionDao) SelectMappingByTrans(trans, syncTrans Transaction) (mapping Mapping, err error) {
	accountType, err := accountModel.NewDao(t.db).GetAccountType(syncTrans.AccountId)
	if err != nil {
		return
	}

	if trans.ID > 0 && syncTrans.ID > 0 {
		switch accountType {
		case accountModel.TypeIndependent:
			err = t.db.Where("main_id = ? AND related_id = ?", syncTrans.ID, trans.ID).First(&mapping).Error
		case accountModel.TypeShare:
			err = t.db.Where("main_id = ? AND related_id = ?", trans.ID, syncTrans.ID).First(&mapping).Error
		default:
			panic("err account.Type")
		}
		return
	}

	if trans.ID > 0 && syncTrans.AccountId > 0 {
		switch accountType {
		case accountModel.TypeIndependent:
			err = t.db.Where("main_account_id = ? AND related_id = ?", syncTrans.AccountId, trans.ID).First(&mapping).Error
		case accountModel.TypeShare:
			err = t.db.Where("main_id = ? AND related_account_id = ?", trans.ID, syncTrans.AccountId).First(&mapping).Error
		default:
			panic("err account.Type")
		}
		return
	}

	if syncTrans.ID > 0 && trans.AccountId > 0 {
		switch accountType {
		case accountModel.TypeIndependent:
			err = t.db.Where("main_id = ? AND related_account_id = ?", syncTrans.ID, trans.AccountId).First(&mapping).Error
		case accountModel.TypeShare:
			err = t.db.Where("main_account_id = ? AND related_id = ?", syncTrans.AccountId, syncTrans.ID).First(&mapping).Error
		default:
			panic("err account.Type")
		}
		return
	}
	err = errors.New("TransactionDao.SelectMappingByTrans query mode is not supported")
	return
}

func (t *TransactionDao) GetAmountRank(accountId uint, ie constant.IncomeExpense, timeCond TimeCondition) (result []Transaction, err error) {
	limit := 10
	query := timeCond.addConditionToQuery(t.db)
	query = query.Where("account_id = ?", accountId).Where("income_expense = ?", ie)
	return result, query.Limit(limit).Order("amount DESC").Find(&result).Error
}
