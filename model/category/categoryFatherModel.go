package categoryModel

import (
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	commonModel "KeepAccount/model/common"
	"database/sql"
	"time"
)

type Father struct {
	ID             uint                   `gorm:"primary_key;column:id;comment:'主键';default:0"`
	AccountID      uint                   `gorm:"column:account_id;index;comment:'账本ID'"`
	IncomeExpense  constant.IncomeExpense `gorm:"column:income_expense;comment:'收支类型'"`
	Name           string                 `gorm:"column:name;size:128;comment:'名称'"`
	Previous       uint                   `gorm:"column:previous;comment:'前一位'"`
	OrderUpdatedAt time.Time
	CreatedAt      time.Time
	commonModel.BaseModel
}

func (f *Father) TableName() string {
	return "category_father"
}

func (f *Father) IsEmpty() bool {
	return f.ID == 0
}

func (f *Father) SelectById(id uint, forUpdate bool) error {
	return commonModel.SelectByIdOfModel(f, id, forUpdate)
}

func (f *Father) Exits(query interface{}, args ...interface{}) (bool, error) {
	return commonModel.ExistOfModel(f, query, args)
}

func (f *Father) CreateOne() error {
	return f.GetDb().Create(f).Error
}

func (f *Father) GetHeader() (*Father, error) {
	result := &Father{}
	err := f.GetDb().Where("previous = 0").First(&result).Error
	return result, err
}

func (f *Father) SetPrevious(previous *Father) error {
	updateData := make(map[string]interface{})
	if previous != nil {
		updateData["previous"] = previous.ID
	} else {
		updateData["previous"] = 0
	}
	updateData["order_updated_at"] = time.Now()
	return f.GetDb().Model(f).Updates(updateData).Error
}

func (f *Father) GetAll(account *accountModel.Account, incomeExpense constant.IncomeExpense) (*sql.Rows, error) {
	db := f.GetDb().Model(&f)
	db.Where("account_id = ? AND income_expense = ?", account.ID, incomeExpense)
	return db.Order("income_expense asc,previous asc,order_updated_at desc").Rows()
}
