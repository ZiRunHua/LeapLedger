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
	Date   time.Time `gorm:"primaryKey;type:TIMESTAMP"`
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

func (s *Statistic) GetDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func ExpenseAccumulate(
	tradeTime time.Time, accountId uint, userId uint, categoryId uint, amount int, count int, tx *gorm.DB,
) error {
	var err error
	var accountSta ExpenseAccountStatistic
	err = accountSta.Accumulate(tradeTime, accountId, amount, count, tx)
	if err != nil {
		return err
	}
	var accountUserSta ExpenseAccountUserStatistic
	err = accountUserSta.Accumulate(tradeTime, accountId, userId, categoryId, amount, count, tx)
	if err != nil {
		return err
	}
	var categorySta ExpenseCategoryStatistic
	err = categorySta.Accumulate(tradeTime, accountId, categoryId, amount, count, tx)
	if err != nil {
		return err
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
		return err
	}
	var accountUserSta IncomeAccountUserStatistic
	err = accountUserSta.Accumulate(tradeTime, accountId, userId, categoryId, amount, count, tx)
	if err != nil {
		return err
	}
	var categorySta IncomeCategoryStatistic
	err = categorySta.Accumulate(tradeTime, accountId, categoryId, amount, count, tx)
	if err != nil {
		return err
	}
	return err
}
