package productModel

import (
	"context"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/global/cus"
	"github.com/ZiRunHua/LeapLedger/global/db"
	"github.com/ZiRunHua/LeapLedger/util/fileTool"
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
	err = db.Transaction(
		context.Background(), func(ctx *cus.TxContext) error {
			tx := ctx.GetDb()
			tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
			return fileTool.ExecSqlFile(sqlFile, tx)
		},
	)
	if err != nil {
		panic(err)
	}
}
