package categoryModel

import (
	"KeepAccount/global"
	"gorm.io/gorm"
)

type categoryDao struct {
	db *gorm.DB
}

func (d *dao) New(db *gorm.DB) *categoryDao {
	if db == nil {
		db = global.GvaDb
	}
	return &categoryDao{db}
}

func (f *categoryDao) GetListByFather(father *Father) ([]Category, error) {
	list := []Category{}
	err := f.db.Where("father_id = ?", father.ID).Find(&list).Error
	return list, err
}
