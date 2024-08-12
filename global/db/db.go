package db

import (
	"KeepAccount/global/cusCtx"
	"KeepAccount/initialize"
	"context"
	"gorm.io/gorm"
)

var (
	db      = initialize.Db
	Context *cusCtx.DbContext
)

func init() {
	Context = cusCtx.WithDb(context.Background(), db)
}

func Get(ctx context.Context) *gorm.DB {
	value := ctx.Value(cusCtx.Db)
	if value == nil {
		return db
	}
	return ctx.Value(cusCtx.Db).(*gorm.DB)
}

type TxFunc func(ctx *cusCtx.TxContext) error

func Transaction(parent context.Context, fc TxFunc) error {
	value := parent.Value(cusCtx.Tx)
	if value != nil {
		return value.(*gorm.DB).Transaction(
			func(tx *gorm.DB) error {
				return fc(cusCtx.WithTx(parent, tx))
			},
		)
	}
	ctx := cusCtx.WithTxCommitContext(parent)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			return fc(cusCtx.WithTx(ctx, tx))
		},
	)
	if err == nil {
		ctx.ExecCallback()
	}
	return err
}

func AddCommitCallback(parent context.Context, callbacks ...cusCtx.TxCommitCallback) error {
	return parent.Value(cusCtx.TxCommit).(*cusCtx.TxCommitContext).AddCallback(callbacks...)
}
