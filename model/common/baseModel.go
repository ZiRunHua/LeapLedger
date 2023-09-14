package commonModel

import (
	"KeepAccount/global"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Model interface {
	GetDb() *gorm.DB
	SetTx(tx *gorm.DB) *gorm.DB
	IsEmpty() bool
}

type BaseModel struct {
	tx *gorm.DB
}

func (base *BaseModel) GetDb() *gorm.DB {
	if base.InTransaction() {
		return base.tx
	} else {
		return global.GvaDb
	}
}

func (base *BaseModel) SetTx(tx *gorm.DB) *gorm.DB {
	base.tx = tx
	return tx
}

func (base *BaseModel) InTransaction() bool {
	if base == nil || base.tx == nil {
		return false
	}
	return true
}

func (base *BaseModel) GetTransaction() *gorm.DB {
	if false == base.InTransaction() {
		panic("not in transaction")
	}
	return base.tx
}

func (base *BaseModel) BeginTransaction() {
	if base.InTransaction() {
		return
	}
	base.SetTx(global.GvaDb.Begin())
}

func (base *BaseModel) CommitTransaction() {
	base.tx.Commit()
}

func (base *BaseModel) DeferCommit(ctx *gin.Context) {
	txn := base.GetTransaction()
	if r := recover(); r != nil {
		// 发生异常时回滚事务
		txn.Rollback()
		panic(r)
	} else if ctx != nil && ctx.IsAborted() {
		// 如果发生Abort，则回滚事务
		txn.Rollback()
	} else {
		// 如果都没有发生，则提交事务
		txn.Commit()
	}
}
