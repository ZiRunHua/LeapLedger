package categoryModel

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	"KeepAccount/util"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CategoryDao struct {
	db *gorm.DB
}

func NewDao(db ...*gorm.DB) *CategoryDao {
	if len(db) > 0 {
		return &CategoryDao{db: db[0]}
	}
	return &CategoryDao{global.GvaDb}
}

// Deprecated: 改用 categoryModel.NewDao
func (d *dao) NewCategory(db *gorm.DB) *CategoryDao {
	if db == nil {
		db = global.GvaDb
	}
	return &CategoryDao{db}
}

func (c *CategoryDao) SelectById(id uint) (category Category, err error) {
	err = c.db.First(&category, id).Error
	return
}

type CategoryUpdateData struct {
	Name *string
	Icon *string
}

func (c *CategoryDao) Update(category Category, data CategoryUpdateData) error {
	updateData := &Category{}
	if err := util.Data.CopyNotEmptyStringOptional(data.Name, &updateData.Name); err != nil {
		return err
	}
	if err := util.Data.CopyNotEmptyStringOptional(data.Icon, &updateData.Icon); err != nil {
		return err
	}
	return c.db.Model(&category).Updates(updateData).Error
}

func (c *CategoryDao) GetListByFather(father *Father) ([]Category, error) {
	list := []Category{}
	err := c.db.Where("father_id = ?", father.ID).Find(&list).Error
	return list, err
}

func (c *CategoryDao) GetListByAccount(account *accountModel.Account) ([]Category, error) {
	list := []Category{}
	err := c.db.Where(
		"account_id = ?", account.ID,
	).Order("income_expense asc,previous asc,order_updated_at desc").Find(&list).Error
	return list, err
}

func (c *CategoryDao) Exist(account accountModel.Account) (bool, error) {
	category := &Category{}
	err := c.db.Where("account_id = ?", account.ID).Take(category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, errors.WithStack(err)
}

func (c *CategoryDao) CreateMapping(parent, child Category) (Mapping, error) {
	mapping := Mapping{
		ParentAccountId:  parent.AccountId,
		ChildAccountId:   child.AccountId,
		ParentCategoryId: parent.ID,
		ChildCategoryId:  child.ID,
	}
	err := c.db.Create(&mapping).Error
	return mapping, err
}

func (c *CategoryDao) GetMappingByAccountMappingOrderByChildCategoryWeight(parentAccountId, childAccountId uint) ([]Mapping, error) {
	query := c.db.Where("category_mapping.parent_account_id = ? AND category_mapping.child_account_id = ?", parentAccountId, childAccountId)
	query = query.Joins("LEFT JOIN category ON category_mapping.child_category_id = category.id")
	var list []Mapping
	err := c.setCategoryOrder(query).Select("category_mapping.*").Find(&list).Error
	return list, err
}

func (c *CategoryDao) setCategoryOrder(db *gorm.DB) *gorm.DB {
	return db.Order("category.income_expense asc,category.previous asc,category.order_updated_at desc")
}
