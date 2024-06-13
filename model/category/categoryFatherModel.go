package categoryModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
	"time"
)

type Father struct {
	ID             uint                   `gorm:"primary_key"`
	AccountId      uint                   `gorm:"index;comment:'账本ID'"`
	IncomeExpense  constant.IncomeExpense `gorm:"comment:'收支类型'"`
	Name           string                 `gorm:"size:128;comment:'名称'"`
	Previous       uint                   `gorm:"comment:'前一位'"`
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
