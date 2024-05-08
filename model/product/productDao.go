package productModel

import (
	"KeepAccount/global"
	"gorm.io/gorm"
)

type ProductDao struct {
	db *gorm.DB
}

func NewDao(db ...*gorm.DB) *ProductDao {
	if len(db) > 0 {
		return &ProductDao{db: db[0]}
	}
	return &ProductDao{global.GvaDb}
}

func (pd *ProductDao) SelectAllCategoryMappingByCategoryId(categoryId uint) (result []TransactionCategoryMapping, err error) {
	err = pd.db.Where("category_id = ?", categoryId).Find(&result).Error
	return
}
