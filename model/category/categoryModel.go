package categoryModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	commonModel "KeepAccount/model/common"
	"gorm.io/gorm"
	"time"
)

type Category struct {
	ID             uint                   `gorm:"primary_key;column:id;comment:'主键'" `
	AccountId      uint                   `gorm:"column:account_id;comment:'账本ID'"`
	FatherId       uint                   `gorm:"column:father_id;comment:'category_father表ID'" `
	IncomeExpense  constant.IncomeExpense `gorm:"column:income_expense;comment:'收支类型'"`
	Name           string                 `gorm:"column:name;size:128;comment:'名称'"`
	Icon           string                 `gorm:"comment:图标;size:64"`
	Previous       uint                   `gorm:"column:previous;comment:'前一位'"`
	OrderUpdatedAt time.Time              `gorm:"default:CURRENT_TIMESTAMP;comment:'顺序更新时间'"`
	CreatedAt      time.Time              `gorm:"default:CURRENT_TIMESTAMP;comment:'创建时间'"`
	commonModel.BaseModel
}

func (c *Category) SelectById(id uint) error {
	return global.GvaDb.First(c, id).Error
}

func (c *Category) GetFather() (father Father, err error) {
	err = global.GvaDb.First(&father, c.FatherId).Error
	return
}

func (c *Category) GetAccount() (result accountModel.Account, err error) {
	err = result.SelectById(c.AccountId)
	return
}

type Mapping struct {
	gorm.Model
	ParentAccountId  uint `gorm:"comment:'父账本ID';index:idx_account_mapping,priority:1"`
	ChildAccountId   uint `gorm:"comment:'子账本ID';index:idx_account_mapping,priority:2" `
	ParentCategoryId uint `gorm:"comment:'父收支类型ID';uniqueIndex:idx_mapping,priority:1"`
	ChildCategoryId  uint `gorm:"comment:'子收支类型ID';uniqueIndex:idx_mapping,priority:2"`
	commonModel.BaseModel
}

func (p *Mapping) TableName() string {
	return "category_mapping"
}

func init() {
	tables := []interface{}{
		Mapping{},
	}
	for _, table := range tables {
		err := global.GvaDb.AutoMigrate(&table)
		if err != nil {
			panic(err)
		}
	}
}
