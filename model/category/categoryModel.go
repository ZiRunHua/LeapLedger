package categoryModel

import (
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	commonModel "KeepAccount/model/common"
	"database/sql"
	"errors"
	"gorm.io/gorm"
	"time"
)

type Category struct {
	ID             uint                   `gorm:"primary_key;column:id;comment:'主键'" `
	AccountID      uint                   `gorm:"column:account_id;comment:'账本ID'"`
	FatherID       uint                   `gorm:"column:father_id;comment:'category_father表ID'" `
	IncomeExpense  constant.IncomeExpense `gorm:"column:income_expense;comment:'收支类型'"`
	Name           string                 `gorm:"column:name;size:128;comment:'名称'"`
	Previous       uint                   `gorm:"column:previous;comment:'前一位'"`
	OrderUpdatedAt time.Time              `gorm:"default:CURRENT_TIMESTAMP;comment:'顺序更新时间'"`
	CreatedAt      time.Time              `gorm:"default:CURRENT_TIMESTAMP;comment:'创建时间'"`
	commonModel.BaseModel
}

func NewCategory(db *gorm.DB) *Category {
	c := new(Category)
	if db != nil {
		c.SetTx(db)
	}
	return c
}
func (c *Category) IsEmpty() bool {
	return c.ID == 0
}
func (c *Category) SelectById(id uint, forUpdate bool) error {
	return commonModel.SelectByIdOfModel(c, id, forUpdate)
}

func (c *Category) Exits(query interface{}, args ...interface{}) (bool, error) {
	return commonModel.ExistOfModel(c, query, args)
}

func (c *Category) GetOneByPrevious(previous uint) error {
	err := c.GetDb().Where("previous = ?", previous).Order("order_updated_at desc").First(&c).Error
	return err
}
func (c *Category) CreateOne() error {
	if c.FatherID == 0 {
		return errors.New("error")
	}
	return c.GetDb().Create(c).Error
}
func (c *Category) GetHead() (*Category, error) {
	result := &Category{}
	db := c.GetDb().Where("account_id = ? AND income_expense = ? AND previous = 0", c.AccountID, c.IncomeExpense)
	err := db.Order("previous asc,order_updated_at desc").First(&result).Error
	return result, err
}
func (c *Category) SetPrevious(previous *Category) error {
	updateData := make(map[string]interface{})
	if previous == nil || previous.IsEmpty() {
		updateData["previous"] = 0
	} else {
		updateData["previous"] = previous.ID
		if c.FatherID != previous.FatherID {
			updateData["father_id"] = previous.FatherID
		}
	}
	updateData["order_updated_at"] = time.Now()
	return c.GetDb().Model(c).Updates(updateData).Error
}
func (c *Category) SetFather() error {
	return c.GetDb().Model(c).Select("father_id", "previous", "order_updated_at").Updates(
		Category{
			FatherID:       c.FatherID,
			Previous:       c.Previous,
			OrderUpdatedAt: time.Now(),
		},
	).Error
}
func (c *Category) GetAll(account *accountModel.Account, incomeExpense *constant.IncomeExpense) (*sql.Rows, error) {
	db := c.GetDb().Model(&c)
	if incomeExpense == nil {
		db.Where("account_id = ?", account.ID)
	} else {
		db.Where("account_id = ? AND income_expense = ?", account.ID, incomeExpense)
	}
	return db.Order("income_expense asc,previous asc,order_updated_at desc").Rows()
}
