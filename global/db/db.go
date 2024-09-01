package db

import (
	"KeepAccount/global/cusCtx"
	"KeepAccount/initialize"
	"context"
	"gorm.io/gorm"
)

var (
	Db      = initialize.Db
	Context *cusCtx.DbContext
)

func init() {
	Context = cusCtx.WithDb(context.Background(), Db)
}

func Get(ctx context.Context) *gorm.DB {
	value := ctx.Value(cusCtx.Db)
	if value == nil {
		return Db
	}
	return ctx.Value(cusCtx.Db).(*gorm.DB)
}

type TxFunc func(ctx *cusCtx.TxContext) error

func Transaction(parent context.Context, fc TxFunc) error {
	ctx := cusCtx.WithTxCommitContext(parent)
	err := Get(ctx).Transaction(
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
