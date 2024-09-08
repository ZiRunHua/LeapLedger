package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	commonModel "KeepAccount/model/common"
	queryFunc "KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	"database/sql"
	"errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Transaction struct {
	ID uint `gorm:"primarykey"`
	Info
	CreatedAt time.Time      `gorm:"type:TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:TIMESTAMP"`
	commonModel.BaseModel
}

type Info struct {
	UserId, AccountId, CategoryId uint
	IncomeExpense                 constant.IncomeExpense
	Amount                        int
	Remark                        string
	TradeTime                     time.Time `gorm:"type:TIMESTAMP"`
}

func (i *Info) Check(db *gorm.DB) error {
	category, err := categoryModel.NewDao(db).SelectById(i.CategoryId)
	if err != nil {
		return err
	}
	accountUser, err := accountModel.NewDao(db).SelectUser(i.AccountId, i.UserId)
	if err != nil {
		return err
	}
	err = accountUser.CheckTransAddByUserId(accountUser.UserId)
	if err != nil {
		return err
	}
	switch true {
	case i.Amount < 0:
		return errors.New("transaction Check:Amount")
	case i.IncomeExpense != category.IncomeExpense:
		return errors.New("transaction Check:IncomeExpense")
	case category.AccountId != i.AccountId:
		return global.ErrAccountId
	}
	return nil
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

func (t *Transaction) Check(db *gorm.DB) error {
	category, err := t.GetCategory(db)
	if err != nil {
		return err
	}
	if category.AccountId != t.AccountId {
		return global.ErrAccountId
	}
	if t.Amount < 0 {
		return errors.New("transaction Check:Amount")
	}
	if t.IncomeExpense != category.IncomeExpense {
		return errors.New("transaction Check:IncomeExpense")
	}
	return nil
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
		Info: Info{
			UserId:        t.UserId,
			IncomeExpense: t.IncomeExpense,
			Amount:        t.Amount,
			Remark:        t.Remark,
			TradeTime:     t.TradeTime,
		},
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
	Location      string
}

func (t *Info) GetStatisticData(isAdd bool) StatisticData {
	if isAdd {
		return StatisticData{
			AccountId: t.AccountId, UserId: t.UserId, IncomeExpense: t.IncomeExpense,
			CategoryId: t.CategoryId, TradeTime: t.TradeTime, Amount: t.Amount, Count: 1,
			Location: accountModel.NewDao().GetLocation(t.AccountId),
		}
	}
	return StatisticData{
		AccountId: t.AccountId, UserId: t.UserId, IncomeExpense: t.IncomeExpense,
		CategoryId: t.CategoryId, TradeTime: t.TradeTime, Amount: -t.Amount, Count: -1,
		Location: accountModel.NewDao().GetLocation(t.AccountId),
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

type Timing struct {
	ID         uint `gorm:"primarykey"`
	AccountId  uint `gorm:"index"`
	UserId     uint
	TransInfo  Info       `gorm:"not null;type:json;serializer:json"`
	Type       TimingType `gorm:"not null;type:char(16)"`
	OffsetDays int        `gorm:"not null;"`
	NextTime   time.Time  `gorm:"not null;"`
	Close      bool
	gorm.Model
}

func (t *Timing) TableName() string { return "transaction_timing" }

type TimingType string

const (
	Once           TimingType = "Once"
	EveryDay       TimingType = "EveryDay"
	EveryWeek      TimingType = "EveryWeek"
	EveryMonth     TimingType = "EveryMonth"
	LastDayOfMonth TimingType = "LastDayOfMonth"
)

func (t *Timing) UpdateNextTime(db *gorm.DB) error {
	var nextTime time.Time
	if t.Type == Once {
		err := db.Model(t).Update("close", true).Error
		if err != nil {
			return err
		}
	} else {
		switch t.Type {
		case EveryDay:
			nextTime = t.NextTime.AddDate(0, 0, 1)
		case EveryWeek:
			nextTime = t.NextTime.AddDate(0, 0, 7)
		case EveryMonth:
			nextTime = t.NextTime.AddDate(0, 1, 0)
		case LastDayOfMonth:
			nextTime = time.Date(t.NextTime.Year(), t.NextTime.Month()+2, 1, 0, 0, 0, 0, time.Local)
			nextTime = nextTime.AddDate(0, 0, -1)
		}
		err := db.Model(t).Updates(map[string]interface{}{
			"trans_info": datatypes.JSONSet("trans_info").Set("trade_time", nextTime),
			"next_time":  datatypes.Date(nextTime),
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Timing) MakeExecTask(db *gorm.DB) (TimingExec, error) {
	exec := TimingExec{
		ConfigId:  t.ID,
		Status:    TimingExecWait,
		TransInfo: t.TransInfo,
	}
	return exec, db.Create(&exec).Error
}

type TimingExec struct {
	ID            uint             `gorm:"primarykey"`
	Status        TimingExecStatus `gorm:"default:0"`
	ConfigId      uint             `gorm:"index;not null"`
	FailCause     string           `gorm:"default:'';not null"`
	TransInfo     Info             `gorm:"not null;type:json;serializer:json"`
	TransactionId uint
	ExecTime      sql.NullTime
	gorm.Model
}
type TimingExecStatus int8

const (
	TimingExecWait TimingExecStatus = iota * 3
	TimingExecFail
	TimingExecSuccess
)

func (t *TimingExec) TableName() string { return "transaction_timing_exec" }
func (t *TimingExec) GetConfig(db *gorm.DB) (Timing, error) {
	return NewDao(db).SelectTimingById(t.ConfigId)
}

func (t *TimingExec) ExecFail(execErr error, db *gorm.DB) error {
	err := db.Model(&Timing{}).Where("id = ?", t.ConfigId).Error
	if err != nil {
		return err
	}
	var failCause string
	if errors.Is(execErr, global.ErrNoPermission) {
		failCause = "账本无权操作"
	}
	return db.Model(t).Where("id = ?", t.ID).Updates(TimingExec{
		FailCause: failCause,
		Status:    TimingExecFail,
		ExecTime:  sql.NullTime{Time: time.Now(), Valid: true},
	}).Error
}

func (t *TimingExec) ExecSuccess(trans Transaction, db *gorm.DB) error {
	return db.Model(t).Where("id = ?", t.ID).Updates(TimingExec{
		TransactionId: trans.ID,
		Status:        TimingExecSuccess,
		ExecTime:      sql.NullTime{Time: time.Now(), Valid: true},
	}).Error
}
