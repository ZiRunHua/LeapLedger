package transaction

import (
	"context"
	"testing"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/cus"
	"github.com/ZiRunHua/LeapLedger/global/db"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	"github.com/stretchr/testify/assert"
)

func TestTransactionDao_CreateAndUpdate(t *testing.T) {

	// 情景1: 正常创建 Transaction 和 Hash
	t.Run(
		"CreateTransactionWithHash", func(t *testing.T) {
			err := db.Transaction(
				context.TODO(), func(ctx *cus.TxContext) error {
					tx, info := ctx.GetDb(), get.TransInfo()
					dao := transactionModel.NewDao(ctx.GetDb())
					transaction, err := dao.Create(info, transactionModel.RecordTypeOfManual)
					assert.NoError(t, err)
					assert.NotZero(t, transaction.ID)

					transaction, err = dao.Create(info, transactionModel.RecordTypeOfManual)
					assert.ErrorIs(t, err, global.ErrTransactionSame)
					assert.NotZero(t, transaction.ID)

					hashValue, err := transaction.Hash()
					assert.NoError(t, err)
					err = tx.Where(
						"account_id = ? AND hash = ?", transaction.AccountId, hashValue,
					).Delete(&transactionModel.Hash{}).Error
					assert.NoError(t, err)

					transaction, err = dao.Create(info, transactionModel.RecordTypeOfManual)
					assert.NoError(t, err)
					assert.NotZero(t, transaction.ID)
					return nil
				},
			)
			assert.NoError(t, err)
		},
	)

	// 情景2: 缺失 Hash 记录时的更新操作
	t.Run(
		"UpdateTransactionWithMissingHash", func(t *testing.T) {
			// 插入交易但删除 Hash 记录
			_ = db.Transaction(
				context.TODO(), func(ctx *cus.TxContext) error {
					tx, info := ctx.GetDb(), get.TransInfo()
					dao := transactionModel.NewDao(tx)
					transaction, err := dao.Create(info, transactionModel.RecordTypeOfManual)
					assert.NoError(t, err)
					assert.NotZero(t, transaction.ID)

					err = tx.Where("trans_id = ?", transaction.ID).Delete(&transactionModel.Hash{}).Error
					assert.NoError(t, err)

					transaction.Remark = t.Name()
					err = dao.Update(&transaction)
					assert.NoError(t, err)

					var hash transactionModel.Hash
					err = tx.Where("trans_id = ?", transaction.ID).First(&hash).Error
					assert.NoError(t, err)
					var updateTrans transactionModel.Transaction
					err = tx.Model(&updateTrans).First(&updateTrans, transaction.ID).Error
					assert.NoError(t, err)
					assert.Equal(t, updateTrans.Info, transaction.Info)
					return nil
				},
			)
		},
	)

	// 情景3: 冲突的 Hash 更新
	t.Run(
		"DuplicateHashConflict", func(t *testing.T) {
			_ = db.Transaction(
				context.TODO(), func(ctx *cus.TxContext) error {
					tx, info := ctx.GetDb(), get.TransInfo()
					dao := transactionModel.NewDao(tx)
					transaction, err := dao.Create(info, transactionModel.RecordTypeOfManual)
					assert.NoError(t, err)
					assert.NotZero(t, transaction.ID)

					transaction.Remark = "Conflict Test"
					hashBytes, err := transaction.Hash()
					assert.NoError(t, err)

					duplicateHash := transactionModel.Hash{
						TransId:   transaction.ID + 1,
						AccountId: info.AccountId,
						Hash:      string(hashBytes),
					}
					err = tx.Create(&duplicateHash).Error
					assert.NoError(t, err)

					err = dao.Update(&transaction)
					assert.ErrorIs(t, err, global.ErrTransactionSame)
					return nil
				},
			)
		},
	)
}
