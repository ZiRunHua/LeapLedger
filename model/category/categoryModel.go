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
	gorm.Model
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

func (c *Category) CheckName(db *gorm.DB) error {
	if c.Name == "" {
		return global.NewErrDataIsEmpty("交易类型名称")
	}
	var exist bool
	sql := "SELECT EXISTS(SELECT 1 FROM category WHERE account_id = ? AND name = ? AND deleted_at is null) AS exist"
	err := db.Raw(sql, c.AccountId, c.Name).Scan(&exist).Error
	if err != nil {
		return err
	} else if exist {
		return global.ErrCategorySameName
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
	gorm.Model
	ParentAccountId  uint `gorm:"comment:'父账本ID';uniqueIndex:idx_mapping,priority:2"`
	ChildAccountId   uint `gorm:"comment:'子账本ID';" `
	ParentCategoryId uint `gorm:"comment:'父收支类型ID';index"`
	ChildCategoryId  uint `gorm:"comment:'子收支类型ID';uniqueIndex:idx_mapping,priority:1"`
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
