package categoryModel

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	"gorm.io/gorm"
)

type fatherDao struct {
	db *gorm.DB
}

func (d *dao) NewFather(db *gorm.DB) *fatherDao {
	if db == nil {
		db = global.GvaDb
	}
	return &fatherDao{db}
}

func (f *fatherDao) GetListByAccount(account *accountModel.Account) ([]Father, error) {
	list := []Father{}
	err := f.db.Where(
		"account_id = ?", account.ID,
	).Order("income_expense asc,previous asc,order_updated_at desc").Find(&list).Error
	return list, err
}
