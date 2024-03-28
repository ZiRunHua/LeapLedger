package categoryModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	commonModel "KeepAccount/model/common"
	queryFunc "KeepAccount/model/common/query"
	"database/sql"
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

func (c *Category) IsEmpty() bool {
	return c.ID == 0
}
func (c *Category) SelectById(id uint) error {
	return global.GvaDb.First(c, id).Error
}

func (c *Category) GetFather() (father Father, err error) {
	err = global.GvaDb.First(&father, c.FatherId).Error
	return
}

func (c *Category) Exits(query interface{}, args ...interface{}) (bool, error) {
	return queryFunc.Exist[*Category](query, args)
}

func (c *Category) GetAccount() (result accountModel.Account, err error) {
	err = result.SelectById(c.AccountId)
	return
}

func (c *Category) GetOneByPrevious(previous uint, tx *gorm.DB) error {
	err := tx.Where("previous = ?", previous).Order("order_updated_at desc").First(&c).Error
	return err
}

func (c *Category) GetHead(tx *gorm.DB) (*Category, error) {
	result := &Category{}
	db := tx.Where("account_id = ? AND income_expense = ? AND previous = 0", c.AccountId, c.IncomeExpense)
	err := db.Order("previous asc,order_updated_at desc").First(&result).Error
	return result, err
}
func (c *Category) SetPrevious(previous *Category, tx *gorm.DB) error {
	updateData := make(map[string]interface{})
	if previous == nil || previous.IsEmpty() {
		updateData["previous"] = 0
	} else {
		updateData["previous"] = previous.ID
		if c.FatherId != previous.FatherId {
			updateData["father_id"] = previous.FatherId
		}
	}
	updateData["order_updated_at"] = time.Now()
	return tx.Model(c).Updates(updateData).Error
}

func (c *Category) GetAll(account accountModel.Account, incomeExpense *constant.IncomeExpense) (*sql.Rows, error) {
	db := global.GvaDb.Model(&c)
	if incomeExpense == nil {
		db.Where("account_id = ?", account.ID)
	} else {
		db.Where("account_id = ? AND income_expense = ?", account.ID, incomeExpense)
	}
	return db.Order("income_expense asc,previous asc,order_updated_at desc").Rows()
}
