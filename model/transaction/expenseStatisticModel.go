package transactionModel

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type ExpenseAccountStatistic struct {
	Statistic
	AccountId uint `gorm:"column:account_id;primaryKey"`
}

func (i *ExpenseAccountStatistic) TableName() string {
	return "transaction_expense_account_statistic"
}

func (e *ExpenseAccountStatistic) Accumulate(
	tradeTime time.Time, accountId uint, amount int, count int, tx *gorm.DB,
) error {
	tradeTime = e.GetDate(tradeTime)
	where := tx.Model(e).Where("date = ? AND account_id = ?", tradeTime, accountId)
	updatesValue := e.GetUpdatesValue(amount, count)
	update := where.Updates(updatesValue)
	err := update.Error

	if update.RowsAffected == 0 {
		e.Date = tradeTime
		e.AccountId = accountId
		e.Amount = amount
		e.Count = count
		err = tx.Create(e).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = where.Updates(updatesValue).Error
		}
	}
	return err
}

type ExpenseAccountUserStatistic struct {
	Statistic
	AccountId uint `gorm:"column:account_id;primaryKey"`
	UserId    uint `gorm:"column:user_id;primaryKey"`
}

func (i *ExpenseAccountUserStatistic) TableName() string {
	return "transaction_expense_account_user_statistic"
}

func (e *ExpenseAccountUserStatistic) Accumulate(
	tradeTime time.Time, accountId uint, userId uint, amount int, count int, tx *gorm.DB,
) error {
	tradeTime = e.GetDate(tradeTime)
	where := tx.Model(e).Where(
		"date = ? AND account_id = ? AND user_id = ?", tradeTime, accountId, userId,
	)
	updatesValue := e.GetUpdatesValue(amount, count)
	update := where.Updates(updatesValue)
	err := update.Error

	if update.RowsAffected == 0 {
		e.Date = tradeTime
		e.AccountId = accountId
		e.UserId = userId
		e.Amount = amount
		e.Count = count
		err = tx.Create(e).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = where.Updates(updatesValue).Error
		}
	}
	return err
}

type ExpenseCategoryStatistic struct {
	Statistic
	CategoryId uint `gorm:"column:category_id;primaryKey"`
	AccountId  uint `gorm:"column:account_id"` //冗余字段
}

func (e *ExpenseCategoryStatistic) TableName() string {
	return "transaction_expense_category_statistic"
}

func (e *ExpenseCategoryStatistic) Accumulate(
	tradeTime time.Time, accountId uint, categoryId uint, amount int, count int, tx *gorm.DB,
) error {
	tradeTime = e.GetDate(tradeTime)
	where := tx.Model(e).Where("date = ? AND category_id = ?", tradeTime, categoryId)
	updatesValue := e.GetUpdatesValue(amount, count)
	update := where.Updates(updatesValue)
	err := update.Error

	if update.RowsAffected == 0 {
		e.Date = tradeTime
		e.CategoryId = categoryId
		e.AccountId = accountId
		e.Amount = amount
		e.Count = count
		err = tx.Create(e).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = where.Updates(updatesValue).Error
		}
	}
	return err
}
