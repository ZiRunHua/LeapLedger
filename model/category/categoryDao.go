package categoryModel

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	"KeepAccount/util"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type categoryDao struct {
	db *gorm.DB
}

func (d *dao) NewCategory(db *gorm.DB) *categoryDao {
	if db == nil {
		db = global.GvaDb
	}
	return &categoryDao{db}
}

type CategoryUpdateData struct {
	Name *string
	Icon *string
}

func (c *categoryDao) Update(category Category, data CategoryUpdateData) error {
	updateData := &Category{}
	if err := util.Data.CopyNotEmptyStringOptional(data.Name, &updateData.Name); err != nil {
		return err
	}
	if err := util.Data.CopyNotEmptyStringOptional(data.Icon, &updateData.Icon); err != nil {
		return err
	}
	return c.db.Model(&category).Updates(updateData).Error
}

func (c *categoryDao) GetListByFather(father *Father) ([]Category, error) {
	list := []Category{}
	err := c.db.Where("father_id = ?", father.ID).Find(&list).Error
	return list, err
}

func (c *categoryDao) GetListByAccount(account *accountModel.Account) ([]Category, error) {
	list := []Category{}
	err := c.db.Where(
		"account_id = ?", account.ID,
	).Order("income_expense asc,previous asc,order_updated_at desc").Find(&list).Error
	return list, err
}

func (c *categoryDao) Exist(account accountModel.Account) (bool, error) {
	category := &Category{}
	err := c.db.Where("account_id = ?", account.ID).Take(category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, errors.WithStack(err)
}
