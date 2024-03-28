package transactionModel

import (
	commonModel "KeepAccount/model/common"
	"gorm.io/gorm"
	"time"
)

type statisticModel interface {
	GetUpdatesValue(amount, count int) map[string]interface{}
	GetDate(tradeTime time.Time) time.Time
	TableName() string
}

type Statistic struct {
	Date   time.Time `gorm:"column:date;primaryKey;type:date"`
	Amount int
	Count  int
	commonModel.BaseModel
}

func (s *Statistic) GetUpdatesValue(amount, count int) map[string]interface{} {
	return map[string]interface{}{
		"amount": gorm.Expr("amount + ?", amount),
		"count":  gorm.Expr("count + ?", count),
	}
}

func (s *Statistic) GetDate(tradeTime time.Time) time.Time {
	return time.Date(tradeTime.Year(), tradeTime.Month(), tradeTime.Day(), 0, 0, 0, 0, time.Local)
}

func ExpenseAccumulate(
	tradeTime time.Time, accountId uint, userId uint, categoryId uint, amount int, count int, tx *gorm.DB,
) error {
	var err error
	var accountSta ExpenseAccountStatistic
	err = accountSta.Accumulate(tradeTime, accountId, amount, count, tx)
	if err != nil {
		return nil
	}
	var accountUserSta ExpenseAccountUserStatistic
	err = accountUserSta.Accumulate(tradeTime, accountId, userId, amount, count, tx)
	if err != nil {
		return nil
	}
	var categorySta ExpenseCategoryStatistic
	err = categorySta.Accumulate(tradeTime, accountId, categoryId, amount, count, tx)
	if err != nil {
		return nil
	}
	return err
}

func IncomeAccumulate(
	tradeTime time.Time, accountId uint, userId uint, categoryId uint, amount int, count int, tx *gorm.DB,
) error {
	var err error
	var accountSta IncomeAccountStatistic
	err = accountSta.Accumulate(tradeTime, accountId, amount, count, tx)
	if err != nil {
		return nil
	}
	var accountUserSta IncomeAccountUserStatistic
	err = accountUserSta.Accumulate(tradeTime, accountId, userId, amount, count, tx)
	if err != nil {
		return nil
	}
	var categorySta IncomeCategoryStatistic
	err = categorySta.Accumulate(tradeTime, accountId, categoryId, amount, count, tx)
	if err != nil {
		return nil
	}
	return err
}
