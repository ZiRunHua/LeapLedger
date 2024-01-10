package transactionModel

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type ExpenseStatistic struct {
	Statistic
}

func (e *ExpenseStatistic) TableName() string {
	return "transaction_expense_statistic"
}

func (e *ExpenseStatistic) Accumulate(
	tradeTime time.Time, categoryId uint, accountId uint, amount int, count int,
) error {
	if amount == 0 {
		return nil
	}
	tradeTime = time.Date(tradeTime.Year(), tradeTime.Month(), tradeTime.Day(), 0, 0, 0, 0, time.Local)
	where := e.GetDb().Model(e).Where("date = ? AND category_id = ?", tradeTime, categoryId).Session(&gorm.Session{})

	var update *gorm.DB
	update = where.Updates(
		map[string]interface{}{
			"amount": gorm.Expr("amount + ?", amount),
			"count":  gorm.Expr("count + ?", count),
		},
	)

	err := update.Error
	if update.RowsAffected == 0 || errors.Is(err, gorm.ErrRecordNotFound) {
		e.Date = tradeTime
		e.CategoryId = categoryId
		e.AccountId = accountId
		e.Amount = amount
		e.Count = count
		err = e.GetDb().Create(e).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			update = where.Updates(
				map[string]interface{}{
					"amount": gorm.Expr("amount + ?", amount),
					"count":  gorm.Expr("count + ?", count),
				},
			)
		}
	}
	return err
}
