package model

import (
	gdb "KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	logModel "KeepAccount/model/log"
	productModel "KeepAccount/model/product"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"context"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	ctx := context.TODO()
	db := gdb.Db.Session(&gorm.Session{Logger: gdb.Db.Logger.LogMode(logger.Silent)})
	errGroup, _ := errgroup.WithContext(ctx)
	errGroup.Go(func() error { return accountModel.CurrentInit(db) })
	errGroup.Go(func() error { return categoryModel.CurrentInit(db) })
	errGroup.Go(func() error { return logModel.CurrentInit(db) })
	errGroup.Go(func() error { return productModel.CurrentInit(db) })
	errGroup.Go(func() error { return transactionModel.CurrentInit(db) })
	errGroup.Go(func() error { return userModel.CurrentInit(db) })
	if err := errGroup.Wait(); err != nil {
		panic(err)
	}
}
