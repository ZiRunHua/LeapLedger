package productModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
)

type TransactionCategory struct {
	ID            uint                   `gorm:"primary_key;column:id"`
	ProductKey    string                 `gorm:"column:product_key"`
	IncomeExpense constant.IncomeExpense `gorm:"column:income_expense;size:8;comment:'收支类型'"`
	Name          string                 `gorm:"column:name;size:64"`
	commonModel.BaseModel
}

func (tc *TransactionCategory) TableName() string {
	return "product_transaction_category"
}

func (tc *TransactionCategory) IsEmpty() bool {
	return tc == nil || tc.ID == 0
}

func (tc *TransactionCategory) SelectById(id uint, forUpdate bool) error {
	return commonModel.SelectByIdOfModel(tc, id, forUpdate)
}

func (tc *TransactionCategory) Exits(query interface{}, args ...interface{}) (bool, error) {
	return commonModel.ExistOfModel(tc, query, args)
}

func (tc *TransactionCategory) GetMap(productKey string) (map[uint]TransactionCategory, error) {
	transCategoryMap := make(map[uint]TransactionCategory)
	var prodTransCategory TransactionCategory
	rows, err := global.GvaDb.Model(&prodTransCategory).Where(
		"product_key = ? ", productKey,
	).Rows()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &prodTransCategory)
		if err != nil {
			return nil, err
		}
		transCategoryMap[prodTransCategory.ID] = prodTransCategory
	}
	return transCategoryMap, nil
}

func (tc *TransactionCategory) GetIncomeExpenseAndNameMap(productKey string) (
	result map[constant.IncomeExpense]map[string]TransactionCategory, err error,
) {
	var prodTransCategory TransactionCategory
	rows, err := global.GvaDb.Model(&prodTransCategory).Where(
		"product_key = ? ", productKey,
	).Rows()
	if err != nil {
		return
	}
	result = map[constant.IncomeExpense]map[string]TransactionCategory{}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &prodTransCategory)
		if err != nil {
			return
		}
		if _, exist := result[prodTransCategory.IncomeExpense]; exist == false {
			result[prodTransCategory.IncomeExpense] = map[string]TransactionCategory{prodTransCategory.Name: prodTransCategory}
		} else {
			result[prodTransCategory.IncomeExpense][prodTransCategory.Name] = prodTransCategory
		}
	}
	return
}
