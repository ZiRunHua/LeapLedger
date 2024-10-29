package db

import (
	"context"
	"github.com/ZiRunHua/LeapLedger/global/cus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
)

func init() {
	Db = Db.Session(&gorm.Session{Logger: Db.Logger.LogMode(logger.Silent)})
}

func callback() {
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i * i
	}
}

func Benchmark_Gorm_ThreeTransaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Db.Transaction(
			func(tx *gorm.DB) error {
				return tx.Transaction(
					func(tx *gorm.DB) error {
						return tx.Transaction(
							func(tx *gorm.DB) error {
								callback()
								callback()
								return nil
							},
						)
					},
				)
			},
		)
	}
}

func Benchmark_Cus_ThreeTransaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Transaction(
			context.Background(), func(ctx *cus.TxContext) error {
				return Transaction(
					ctx, func(ctx *cus.TxContext) error {
						return Transaction(
							ctx, func(ctx *cus.TxContext) error {
								_ = AddCommitCallback(ctx, callback, callback)
								return nil
							},
						)
					},
				)
			},
		)
	}
}
func Benchmark_Gorm_Transaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Db.Transaction(
			func(tx *gorm.DB) error {
				callback()
				callback()
				return nil
			},
		)
	}
}

func Benchmark_Cus_Transaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Transaction(
			context.Background(), func(ctx *cus.TxContext) error {
				_ = AddCommitCallback(ctx, callback, callback)
				return nil
			},
		)
	}
}
