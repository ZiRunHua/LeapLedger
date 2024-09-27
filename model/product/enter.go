package productModel

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/util/fileTool"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

var initSqlFile = constant.DATA_PATH + "/database/product.sql"

func init() {
	// table
	tables := []interface{}{
		Product{}, BillHeader{}, Bill{},
		TransactionCategory{}, TransactionCategoryMapping{},
	}
	err := db.InitDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
	// table data
	sqlFile, err := os.Open(initSqlFile)
	if err != nil {
		panic(err)
	}
	err = db.Transaction(context.Background(), func(ctx *cus.TxContext) error {
		tx := ctx.GetDb()
		tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
		return fileTool.ExecSqlFile(sqlFile, tx)
	})
	if err != nil {
		panic(err)
	}
}
