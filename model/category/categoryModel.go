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
	ID             uint                   `gorm:"comment:'主键';primary_key;" `
	AccountId      uint                   `gorm:"comment:'账本ID';uniqueIndex:unique_name,priority:1"`
	FatherId       uint                   `gorm:"comment:'category_father表ID';index" `
	IncomeExpense  constant.IncomeExpense `gorm:"comment:'收支类型'"`
	Name           string                 `gorm:"comment:'名称';size:128;uniqueIndex:unique_name,priority:2"`
	Icon           string                 `gorm:"comment:'图标';size:64"`
	Previous       uint                   `gorm:"comment:'前一位'"`
	OrderUpdatedAt time.Time              `gorm:"comment:'顺序更新时间';not null;default:now();type:TIMESTAMP;"`
	CreatedAt      time.Time              `gorm:"type:TIMESTAMP"`
	UpdatedAt      time.Time              `gorm:"type:TIMESTAMP"`
	DeletedAt      gorm.DeletedAt         `gorm:"index;type:TIMESTAMP"`
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

func (c *Category) CheckName(_ *gorm.DB) error {
	if c.Name == "" {
		return global.NewErrDataIsEmpty("交易类型名称")
	}
	return nil
}

type Condition struct {
	account accountModel.Account
	ie      *constant.IncomeExpense
}

func (c *Condition) buildWhere(db *gorm.DB) *gorm.DB {
	if c.ie == nil {
		return db.Where("account_id = ?", c.account.ID)
	}
	return db.Where("account_id = ? AND income_expense = ?", c.account.ID, c.ie)
}

// Mapping
// ParentAccountId - ChildCategoryId unique
// ParentCategoryId - ChildCategoryId  unique
// ChildAccountId - ParentCategoryId  unique
type Mapping struct {
	ID               uint           `gorm:"primarykey"`
	ParentAccountId  uint           `gorm:"comment:'父账本ID';uniqueIndex:idx_mapping,priority:2"`
	ChildAccountId   uint           `gorm:"comment:'子账本ID';" `
	ParentCategoryId uint           `gorm:"comment:'父收支类型ID';index"`
	ChildCategoryId  uint           `gorm:"comment:'子收支类型ID';uniqueIndex:idx_mapping,priority:1"`
	CreatedAt        time.Time      `gorm:"type:TIMESTAMP"`
	UpdatedAt        time.Time      `gorm:"type:TIMESTAMP"`
	DeletedAt        gorm.DeletedAt `gorm:"index;type:TIMESTAMP"`
	commonModel.BaseModel
}

func (p *Mapping) TableName() string {
	return "category_mapping"
}
