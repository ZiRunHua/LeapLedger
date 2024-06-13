package productModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
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
func (pd *ProductDao) SelectByName(key KeyValue, ie constant.IncomeExpense, name string) (result TransactionCategory, err error) {
	err = pd.db.Where("product_key = ? AND income_expense = ? AND name = ?", key, ie, name).First(&result).Error
	return
}

func (pd *ProductDao) SelectAllCategoryMappingByCategoryId(categoryId uint) (result []TransactionCategoryMapping, err error) {
	err = pd.db.Where("category_id = ?", categoryId).Find(&result).Error
	return
}
