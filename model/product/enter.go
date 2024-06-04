package productModel

import (
	"KeepAccount/global"
)

func init() {
	tables := []interface{}{
		Product{}, BillHeader{}, Bill{},
		TransactionCategory{}, TransactionCategoryMapping{},
	}
	err := global.GvaDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
