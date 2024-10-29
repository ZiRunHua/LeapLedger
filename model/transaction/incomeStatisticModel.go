package transactionModel

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type IncomeAccountStatistic struct {
	Statistic
	AccountId uint `gorm:"primaryKey"`
}

func (i *IncomeAccountStatistic) TableName() string {
	return "transaction_income_account_statistic"
}

func (i *IncomeAccountStatistic) Accumulate(
	tradeTime time.Time, accountId uint, amount int, count int, tx *gorm.DB,
) error {
	tradeTime = i.GetDate(tradeTime)
	where := tx.Model(i).Where("date = ? AND account_id = ?", tradeTime, accountId)
	updatesValue := i.GetUpdatesValue(amount, count)
	update := where.Updates(updatesValue)
	err := update.Error
	if err != nil {
		return err
	}
	if update.RowsAffected == 0 {
		i.Date = tradeTime
		i.AccountId = accountId
		i.Amount = amount
		i.Count = count
		err = tx.Create(i).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = where.Updates(updatesValue).Error
		}
	}
	return err
}

type IncomeAccountUserStatistic struct {
	Statistic
	AccountId  uint `gorm:"primaryKey"`
	UserId     uint `gorm:"primaryKey"`
	CategoryId uint `gorm:"primaryKey"`
}

func (i *IncomeAccountUserStatistic) TableName() string {
	return "transaction_income_account_user_statistic"
}

func (i *IncomeAccountUserStatistic) Accumulate(
	tradeTime time.Time, accountId uint, userId uint, categoryId uint, amount int, count int, tx *gorm.DB,
) error {
	tradeTime = i.GetDate(tradeTime)
	where := tx.Model(i).Where(
		"date = ? AND account_id = ? AND user_id = ? AND category_id = ?", tradeTime, accountId, userId, categoryId,
	)
	updatesValue := i.GetUpdatesValue(amount, count)
	update := where.Updates(updatesValue)
	err := update.Error
	if err != nil {
		return err
	}
	if update.RowsAffected == 0 {
		i.Date = tradeTime
		i.AccountId = accountId
		i.UserId = userId
		i.CategoryId = categoryId
		i.Amount = amount
		i.Count = count
		err = tx.Create(i).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = where.Updates(updatesValue).Error
		}
	}
	return err
}

type IncomeCategoryStatistic struct {
	CategoryId uint `gorm:"primaryKey"`
	AccountId  uint //冗余字段
	Statistic
}

func (i *IncomeCategoryStatistic) TableName() string {
	return "transaction_income_category_statistic"
}

func (i *IncomeCategoryStatistic) Accumulate(
	tradeTime time.Time, accountId uint, categoryId uint, amount int, count int, tx *gorm.DB,
) error {
	tradeTime = i.GetDate(tradeTime)
	where := tx.Model(i).Where("date = ? AND category_id = ?", tradeTime, categoryId)
	updatesValue := i.GetUpdatesValue(amount, count)
	update := where.Updates(updatesValue)
	err := update.Error
	if err != nil {
		return err
	}
	if update.RowsAffected == 0 {
		i.Date = tradeTime
		i.CategoryId = categoryId
		i.AccountId = accountId
		i.Amount = amount
		i.Count = count
		err = tx.Create(i).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = where.Updates(updatesValue).Error
		}
	}
	return err
}
