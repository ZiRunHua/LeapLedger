package model

import (
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	logModel "KeepAccount/model/log"
	productModel "KeepAccount/model/product"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"context"
	"golang.org/x/sync/errgroup"
)

func init() {
	errGroup, _ := errgroup.WithContext(context.TODO())
	errGroup.Go(func() error { return accountModel.CurrentInit() })
	errGroup.Go(func() error { return categoryModel.CurrentInit() })
	errGroup.Go(func() error { return logModel.CurrentInit() })
	errGroup.Go(func() error { return productModel.CurrentInit() })
	errGroup.Go(func() error { return transactionModel.CurrentInit() })
	errGroup.Go(func() error { return userModel.CurrentInit() })
	if err := errGroup.Wait(); err != nil {
		panic(err)
	}
}
