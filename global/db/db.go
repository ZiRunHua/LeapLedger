package db

import (
	"KeepAccount/global/cus"
	"KeepAccount/initialize"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Db      = initialize.Db
	InitDb  = Db.Session(&gorm.Session{Logger: Db.Logger.LogMode(logger.Silent)})
	Context *cus.DbContext
)

func init() {
	Context = cus.WithDb(context.Background(), Db)
}

func Get(ctx context.Context) *gorm.DB {
	value := ctx.Value(cus.Db)
	if value == nil {
		return Db
	}
	return ctx.Value(cus.Db).(*gorm.DB)
}

type TxFunc func(ctx *cus.TxContext) error

func Transaction(parent context.Context, fc TxFunc) error {
	ctx := cus.WithTxCommitContext(parent)
	err := Get(ctx).Transaction(
		func(tx *gorm.DB) error {
			return fc(cus.WithTx(ctx, tx))
		},
	)
	if err == nil {
		ctx.ExecCallback()
	}
	return err
}

func AddCommitCallback(parent context.Context, callbacks ...cus.TxCommitCallback) error {
	return parent.Value(cus.TxCommit).(*cus.TxCommitContext).AddCallback(callbacks...)
}
