package transactionModel

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type ExpenseAccountStatistic struct {
	Statistic
	AccountId uint `gorm:"primaryKey"`
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
	if err != nil {
		return err
	}
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
	AccountId  uint `gorm:"primaryKey"`
	UserId     uint `gorm:"primaryKey"`
	CategoryId uint `gorm:"primaryKey"`
}

func (i *ExpenseAccountUserStatistic) TableName() string {
	return "transaction_expense_account_user_statistic"
}

func (e *ExpenseAccountUserStatistic) Accumulate(
	tradeTime time.Time, accountId uint, userId uint, categoryId uint, amount int, count int, tx *gorm.DB,
) error {
	tradeTime = e.GetDate(tradeTime)
	where := tx.Model(e).Where(
		"date = ? AND account_id = ? AND user_id = ? AND category_id = ?", tradeTime, accountId, userId, categoryId,
	)
	updatesValue := e.GetUpdatesValue(amount, count)
	update := where.Updates(updatesValue)
	err := update.Error
	if err != nil {
		return err
	}
	if update.RowsAffected == 0 {
		e.Date = tradeTime
		e.AccountId = accountId
		e.UserId = userId
		e.CategoryId = categoryId
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
	CategoryId uint `gorm:"primaryKey"`
	AccountId  uint //冗余字段
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
	if err != nil {
		return err
	}
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
