package transactionModel

import (
	"database/sql"
	"time"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	"github.com/ZiRunHua/LeapLedger/util/timeTool"

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

func (t *TransactionDao) Create(info Info, recordType RecordType) (result Transaction, err error) {
	result.Info, result.RecordType = info, recordType
	return result, t.db.Create(&result).Error
}

func (t *TransactionDao) GetListByCondition(condition Condition, offset int, limit int) (
	result []Transaction, err error,
) {
	query := condition.addConditionToQuery(t.db)
	err = query.Limit(limit).Offset(offset).Order("trade_time DESC").Find(&result).Error
	return
}

func (t *TransactionDao) GetIeStatisticByCondition(
	ie *constant.IncomeExpense, condition StatisticCondition, extCond *ExtensionCondition,
) (result global.IEStatistic, err error) {
	if extCond.IsSet() {
		// transaction table select
		query := t.db.Model(&Transaction{})
		query = condition.ForeignKeyCondition.addConditionToQuery(query)
		query, err = t.setTimeRangeForQuery(
			query, timeTool.ToDay(condition.StartTime), timeTool.ToDay(condition.EndTime),
		)
		if err != nil {
			return
		}
		query = extCond.addConditionToQuery(query)
		result, err = t.getIEStatisticByWhere(ie, query)
	} else {
		// statistic table select
		result, err = NewStatisticDao(t.db).GetIeStatisticByCondition(ie, condition)
	}
	if err != nil {
		err = errors.Wrap(err, "transactionDao.GetIeStatisticByCondition")
	}
	return
}

func (t *TransactionDao) setTimeRangeForQuery(query *gorm.DB, startTime, endTime time.Time) (*gorm.DB, error) {
	switch true {
	case !startTime.IsZero() && !endTime.IsZero():
		query = query.Where("trade_time BETWEEN ? AND ?", startTime, endTime)
	case !startTime.IsZero():
		query = query.Where("trade_time >=", startTime)
	case !endTime.IsZero():
		query = query.Where("trade_time <=", endTime)
	}
	return query, nil
}

func (t *TransactionDao) getIEStatisticByWhere(ie *constant.IncomeExpense, query *gorm.DB) (
	result global.IEStatistic, err error,
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
			panic(accountModel.ErrAccountType)
		}
		return
	}

	if trans.ID > 0 && syncTrans.AccountId > 0 {
		switch accountType {
		case accountModel.TypeIndependent:
			err = t.db.Where(
				"main_account_id = ? AND related_id = ?", syncTrans.AccountId,
				trans.ID,
			).First(&mapping).Error
		case accountModel.TypeShare:
			err = t.db.Where(
				"main_id = ? AND related_account_id = ?", trans.ID,
				syncTrans.AccountId,
			).First(&mapping).Error
		default:
			panic("err account.Type")
		}
		return
	}

	if syncTrans.ID > 0 && trans.AccountId > 0 {
		switch accountType {
		case accountModel.TypeIndependent:
			err = t.db.Where(
				"main_id = ? AND related_account_id = ?", syncTrans.ID,
				trans.AccountId,
			).First(&mapping).Error
		case accountModel.TypeShare:
			err = t.db.Where(
				"main_account_id = ? AND related_id = ?", syncTrans.AccountId,
				syncTrans.ID,
			).First(&mapping).Error
		default:
			panic("err account.Type")
		}
		return
	}
	err = errors.New("TransactionDao.SelectMappingByTrans query mode is not supported")
	return
}

func (t *TransactionDao) GetAmountRank(
	accountId uint, ie constant.IncomeExpense, timeCond TimeCondition,
) (result []Transaction, err error) {
	limit := 10
	query := timeCond.addConditionToQuery(t.db)
	query = query.Where("account_id = ?", accountId).Where("income_expense = ?", ie)
	return result, query.Limit(limit).Order("amount DESC").Find(&result).Error
}

func (t *TransactionDao) SelectTimingById(id uint) (result Timing, err error) {
	err = t.db.First(&result, id).Error
	return
}

func (t *TransactionDao) SelectTimingListByUserId(accountId uint, offset int, limit int) (result []Timing, err error) {
	err = t.db.Where("account_id = ?", accountId).Limit(limit).Offset(offset).Order("id DESC").Find(&result).Error
	return
}

func (t *TransactionDao) SelectAllTimingAndProcess(startTime time.Time, process func(timing Timing) error) (err error) {
	rows, err := t.db.Model(&Timing{}).Where("next_time < ? AND close = ?", startTime, false).Rows()
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		if err != nil {
			_ = rows.Close()
		} else {
			err = rows.Close()
		}
	}(rows)
	var timing Timing
	for rows.Next() {
		err = t.db.ScanRows(rows, &timing)
		if err != nil {
			return err
		}
		err = process(timing)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TransactionDao) SelectWaitTimingExec(startId uint, limit int) ([]TimingExec, error) {
	var list []TimingExec
	err := t.db.Where("id >= ? AND status = ?", startId, TimingExecWait).Order("id ASC").Limit(limit).Find(&list).Error
	return list, err
}
