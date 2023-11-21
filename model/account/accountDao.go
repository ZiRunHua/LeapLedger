package accountModel

import (
	"KeepAccount/global"
	"KeepAccount/util"
	"gorm.io/gorm"
)

type accountDao struct {
	db *gorm.DB
}

func (*dao) NewAccount(db *gorm.DB) *accountDao {
	if db == nil {
		db = global.GvaDb
	}
	return &accountDao{db}
}

type AccountUpdateData struct {
	Name *string
	Icon *string
}

func (a *accountDao) Update(account *Account, data *AccountUpdateData) error {
	updateData := &Account{}
	if err := util.Data.CopyNotEmptyStringOptional(data.Name, &updateData.Name); err != nil {
		return err
	}
	if err := util.Data.CopyNotEmptyStringOptional(data.Icon, &updateData.Icon); err != nil {
		return err
	}
	return a.db.Model(&account).Updates(updateData).Error
}
