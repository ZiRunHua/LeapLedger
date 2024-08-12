package productModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/util"
	"KeepAccount/util/gormFunc"
	"gorm.io/gorm"
	"os"
)

var initSqlFile = constant.DATA_PATH + "/database/product.sql"

func CurrentInit(db *gorm.DB) error {
	// table
	tables := []interface{}{
		Product{}, BillHeader{}, Bill{},
		TransactionCategory{}, TransactionCategoryMapping{},
	}
	err := db.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		_ = gormFunc.AlterIdToHeader(table, db)
	}
	// table data
	sqlFile, err := os.Open(initSqlFile)
	if err != nil {
		return err
	}
	return global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return util.File.ExecSqlFile(sqlFile, tx)
		},
	)
}
