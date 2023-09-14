package categoryModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
	"time"
)

type Father struct {
	ID             uint                   `gorm:"primary_key;column:id;comment:'主键';default:0"`
	AccountId      uint                   `gorm:"column:account_id;index;comment:'账本ID'"`
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

func (f *Father) SelectById(id uint) error {
	return global.GvaDb.First(f, id).Error
}
