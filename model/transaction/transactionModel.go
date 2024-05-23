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

func (t *Transaction) ForShare(tx *gorm.DB) error {
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).First(t).Error; err != nil {
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
// MainId - RelatedId unique
// MainId - RelatedAccountId unique
type Mapping struct {
	ID               uint `gorm:"primarykey"`
	MainId           uint `gorm:"not null;uniqueIndex:idx_mapping,priority:1"`
	MainAccountId    uint `gorm:"not null;"`
	RelatedId        uint `gorm:"not null;"`
	RelatedAccountId uint `gorm:"not null;uniqueIndex:idx_mapping,priority:2"`
	// 上次引起同步的交易更新时间，用来避免错误重试导致旧同步覆盖新同步
	LastSyncedTransUpdateTime time.Time `gorm:"not null;comment:'上次引起同步的交易更新时间'"`
	gorm.Model
}

func (m *Mapping) TableName() string { return "transaction_mapping" }

func (m *Mapping) ForShare(tx *gorm.DB) error {
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).First(m).Error; err != nil {
		return err
	}
	return nil
}

func (m *Mapping) CanSyncTrans(transaction Transaction) bool {
	return transaction.UpdatedAt.After(m.LastSyncedTransUpdateTime)
}

func (m *Mapping) OnSyncSuccess(db *gorm.DB, transaction Transaction) error {
	return db.Model(m).Update("last_synced_trans_update_time", transaction.UpdatedAt).Error
}
