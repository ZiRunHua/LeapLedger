package productModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
)

type TransactionCategory struct {
	ID            uint                   `gorm:"primary_key"`
	ProductKey    string                 `gorm:"uniqueIndex:unique_name,priority:1"`
	IncomeExpense constant.IncomeExpense `gorm:"size:8;comment:'收支类型';uniqueIndex:unique_name,priority:2"`
	Name          string                 `gorm:"size:64;uniqueIndex:unique_name,priority:3"`
	commonModel.BaseModel
}

func (tc *TransactionCategory) TableName() string {
	return "product_transaction_category"
}

func (tc *TransactionCategory) IsEmpty() bool {
	return tc == nil || tc.ID == 0
}

func (tc *TransactionCategory) SelectById(id uint) error {
	return global.GvaDb.First(tc, id).Error
}

func (tc *TransactionCategory) GetMap(productKey KeyValue) (map[uint]TransactionCategory, error) {
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

func (tc *TransactionCategory) GetIncomeExpenseAndNameMap(productKey KeyValue) (
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
