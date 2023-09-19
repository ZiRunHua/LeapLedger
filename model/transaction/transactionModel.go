package transaction

import (
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
	"gorm.io/gorm"
	"time"
)

type Transaction struct {
	gorm.Model
	AccountID     uint `gorm:"column:account_id"`
	CategoryID    uint `gorm:"column:category_id"`
	IncomeExpense constant.IncomeExpense
	Amount        int
	Remark        string
	TradeTime     time.Time
	commonModel.BaseModel
}

func NewTransaction(db *gorm.DB) *Transaction {
	t := &Transaction{}
	t.SetTx(db)
	return t
}
func (t *Transaction) IsEmpty() bool {
	return t.ID == 0
}

func (t *Transaction) SelectById(id uint, forUpdate bool) error {
	return commonModel.SelectByIdOfModel(t, id, forUpdate)
}

func (t *Transaction) Exits(query interface{}, args ...interface{}) (bool, error) {
	return commonModel.ExistOfModel(t, query, args...)
}

func (t *Transaction) Update() error {
	return t.GetDb().Updates(t).Error
}

func (t *Transaction) CreateOne(transaction *Transaction) error {
	return t.GetDb().Create(transaction).Error
}
