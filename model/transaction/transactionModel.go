package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	commonModel "KeepAccount/model/common"
	queryFunc "KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Transaction struct {
	gorm.Model
	UserId        uint `gorm:"column:user_id"`
	AccountId     uint `gorm:"column:account_id"`
	CategoryId    uint `gorm:"column:category_id"`
	IncomeExpense constant.IncomeExpense
	Amount        int
	Remark        string
	TradeTime     time.Time
	commonModel.BaseModel
}

func (t *Transaction) ForUpdate(tx *gorm.DB) error {
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Transaction) SelectById(id uint) error {
	return global.GvaDb.First(t, id).Error
}

func (t *Transaction) Exits(query interface{}, args ...interface{}) (bool, error) {
	return queryFunc.Exist[*Transaction](query, args)
}

func (t *Transaction) GetCategory(db ...*gorm.DB) (category categoryModel.Category, err error) {
	if len(db) > 0 {
		err = db[0].First(&category, t.CategoryId).Error
	} else {
		err = global.GvaDb.First(&category, t.CategoryId).Error
	}
	return
}

func (t *Transaction) GetUser(selects ...interface{}) (user userModel.User, err error) {
	if len(selects) > 0 {
		err = global.GvaDb.Select(selects[0], selects[1:]...).First(&user, t.UserId).Error
	} else {
		err = global.GvaDb.First(&user, t.UserId).Error
	}
	return
}

func (t *Transaction) GetAccount(db ...*gorm.DB) (account accountModel.Account, err error) {
	if len(db) > 0 {
		err = db[0].First(&account, t.AccountId).Error
	} else {
		err = global.GvaDb.First(&account, t.AccountId).Error
	}
	return
}

func (t *Transaction) SyncDataClone() Transaction {
	return Transaction{
		UserId:        t.UserId,
		IncomeExpense: t.IncomeExpense,
		Amount:        t.Amount,
		Remark:        t.Remark,
		TradeTime:     t.TradeTime,
	}
}

type StatisticData struct {
	AccountId     uint
	UserId        uint
	IncomeExpense constant.IncomeExpense
	CategoryId    uint
	TradeTime     time.Time
	Amount        int
	Count         int
}

func (t *Transaction) GetStatisticData(isAdd bool) StatisticData {
	if isAdd {
		return StatisticData{
			AccountId: t.AccountId, UserId: t.UserId, IncomeExpense: t.IncomeExpense,
			CategoryId: t.CategoryId, TradeTime: t.TradeTime, Amount: t.Amount, Count: 1,
		}
	}
	return StatisticData{
		AccountId: t.AccountId, UserId: t.UserId, IncomeExpense: t.IncomeExpense,
		CategoryId: t.CategoryId, TradeTime: t.TradeTime, Amount: -t.Amount, Count: -1,
	}
}

// Mapping
// 一个 MainId 对应多个 RelatedId  因为一笔交易可能同步到多个账本 同时 MainId 和 RelatedId 唯一
// MainId 和 RelatedAccountId 唯一  因为一笔交易只会被同步一个账本一次
type Mapping struct {
	ID               uint `gorm:"primarykey"`
	MainId           uint `gorm:"not null;uniqueIndex:idx_mapping,priority:1"`
	MainAccountId    uint `gorm:"not null;"`
	RelatedId        uint `gorm:"not null;"`
	RelatedAccountId uint `gorm:"not null;uniqueIndex:idx_mapping,priority:2"`
	gorm.Model
}
