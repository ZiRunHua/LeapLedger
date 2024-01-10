package transactionModel

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type IncomeStatistic struct {
	Statistic
}

func (i *IncomeStatistic) TableName() string {
	return "transaction_income_statistic"
}
func (i *IncomeStatistic) Accumulate(
	tradeTime time.Time, categoryId uint, accountId uint, amount int, count int,
) error {
	if amount == 0 {
		return nil
	}
	tradeTime = time.Date(tradeTime.Year(), tradeTime.Month(), tradeTime.Day(), 0, 0, 0, 0, time.Local)
	where := i.GetDb().Model(i).Where("date = ? AND category_id = ?", tradeTime, categoryId).Session(&gorm.Session{})

	var update *gorm.DB
	update = where.Updates(
		map[string]interface{}{
			"amount": gorm.Expr("amount + ?", amount),
			"count":  gorm.Expr("count + ?", count),
		},
	)
	err := update.Error
	if update.RowsAffected == 0 || errors.Is(err, gorm.ErrRecordNotFound) {
		i.Date = tradeTime
		i.CategoryId = categoryId
		i.AccountId = accountId
		i.Amount = amount
		i.Count = count
		err = i.GetDb().Create(i).Error
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
