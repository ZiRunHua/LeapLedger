package transaction

import (
	commonModel "KeepAccount/model/common"
	"errors"
	"gorm.io/gorm"
	"time"
)

type ExpenseStatistic struct {
	Date       time.Time `gorm:"column:date;primaryKey" json:"date"`
	CategoryID uint      `gorm:"column:category_id;primaryKey" json:"category_id"`
	AccountID  uint      `gorm:"column:account_id;" json:"account_id"` //冗余字段
	Amount     int       `gorm:"column:amount" json:"amount"`
	commonModel.BaseModel
}

func (e *ExpenseStatistic) TableName() string {
	return "transaction_expense_statistic"
}

func (i *ExpenseStatistic) Accumulate(
	tradeTime time.Time, categoryId uint, accountId uint, amount int,
) error {
	if amount == 0 {
		return nil
	}
	where := i.GetDb().Model(i).Where("date = ? AND category_id = ?", tradeTime, categoryId)
	result := where.Update("amount", gorm.Expr("amount + ?", amount))
	err := result.Error
	if result.RowsAffected == 0 || errors.Is(err, gorm.ErrRecordNotFound) {
		i.Date = tradeTime
		i.CategoryID = categoryId
		i.AccountID = accountId
		i.Amount = amount
		err = i.GetDb().Create(i).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = where.Update("amount", gorm.Expr("amount + ?", amount)).Error
		}
	}
	return err
}
