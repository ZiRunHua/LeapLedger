package cus

import (
	"context"
	"gorm.io/gorm"
)

func WithDb(parent context.Context, db *gorm.DB) *DbContext {
	return &DbContext{Context: parent, db: db}
}

type DbContext struct {
	context.Context
	db *gorm.DB
}

func (dc *DbContext) Value(key any) any {
	if key == Db {
		return dc.db
	}
	return dc.Context.Value(key)
}
func (dc *DbContext) GetDb() *gorm.DB {
	return dc.db
}

func WithTx(parent context.Context, tx *gorm.DB) *TxContext {
	return &TxContext{Context: parent, tx: tx}
}

type TxContext struct {
	context.Context
	tx *gorm.DB
}

func (tc *TxContext) Value(key any) any {
	if key == Db || key == Tx {
		return tc.tx
	}
	return tc.Context.Value(key)
}

func (tc *TxContext) GetDb() *gorm.DB {
	return tc.tx
}

func WithTxCommitContext(parent context.Context) *TxCommitContext {
	return &TxCommitContext{Context: parent}
}

type TxCommitCallback func()

type TxCommitContext struct {
	context.Context
	callbacks []TxCommitCallback
}

func (t *TxCommitContext) Value(key any) any {
	if key == TxCommit {
		return t
	}
	return t.Context.Value(key)
}

func (t *TxCommitContext) AddCallback(callback ...TxCommitCallback) error {
	t.callbacks = append(t.callbacks, callback...)
	return nil
}

func (t *TxCommitContext) ExecCallback() {
	if len(t.callbacks) == 0 {
		return
	}
	parent := t.Context.Value(TxCommit)
	if parent != nil {
		// The parent transaction decides to commit last, so the callback is handed over to the parent transaction
		err := parent.(*TxCommitContext).AddCallback(t.callbacks...)
		if err != nil {
			panic(err)
		}
		return
	}
	for _, callback := range t.callbacks {
		callback()
	}
}
