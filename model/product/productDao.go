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
func (pd *ProductDao) SelectByKey(key Key) (result Product, err error) {
	return result, pd.db.Where("`key` = ?", key).First(&result).Error
}

func (pd *ProductDao) SelectBillByKey(key Key) (result Bill, err error) {
	return result, pd.db.Where("`product_key` = ?", key).First(&result).Error
}
func (pd *ProductDao) SelectCategoryByName(
	key Key, ie constant.IncomeExpense, name string,
) (result TransactionCategory, err error) {
	err = pd.db.Where("product_key = ? AND income_expense = ? AND name = ?", key, ie, name).First(&result).Error
	return
}

func (pd *ProductDao) SelectAllCategoryMappingByCategoryId(categoryId uint) (
	result []TransactionCategoryMapping, err error,
) {
	err = pd.db.Where("category_id = ?", categoryId).Find(&result).Error
	return
}
func (pd *ProductDao) GetPtcIdMapping(accountId uint, productKey Key) (
	result map[uint]TransactionCategoryMapping, err error,
) {
	var list []TransactionCategoryMapping
	err = pd.db.Where("account_id = ? AND product_key = ?", accountId, productKey).Find(&list).Error
	if err != nil {
		return
	}
	result = make(map[uint]TransactionCategoryMapping)
	for _, mapping := range list {
		result[mapping.PtcId] = mapping
	}
	return
}

func (pd *ProductDao) GetIncomeExpenseAndNameMap(productKey Key) (
	result map[constant.IncomeExpense]map[string]TransactionCategory, err error,
) {
	var list []TransactionCategory
	err = pd.db.Where("product_key = ? ", productKey).Find(&list).Error
	if err != nil {
		return
	}
	result = make(map[constant.IncomeExpense]map[string]TransactionCategory)
	for _, category := range list {
		if _, exist := result[category.IncomeExpense]; exist == false {
			result[category.IncomeExpense] = map[string]TransactionCategory{category.Name: category}
		} else {
			result[category.IncomeExpense][category.Name] = category
		}
	}
	return
}
func (pd *ProductDao) GetBillHeaderList(productKey Key) (list []BillHeader, err error) {
	err = pd.db.Where("product_key = ? ", productKey).Order("id ASC").Find(&list).Error
	return
}
