package transaction

import (
	"context"
	"fmt"
	"testing"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/cus"
	"github.com/ZiRunHua/LeapLedger/global/db"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	"github.com/stretchr/testify/assert"
)

func TestTransactionDao_CreateAndUpdate(t *testing.T) {
	// Scenario 1: A Transaction and Hash are created normally
	t.Run(
		"CreateTransactionWithHash", func(t *testing.T) {
			err := db.Transaction(
				context.TODO(), func(ctx *cus.TxContext) error {
					tx, info := ctx.GetDb(), get.TransInfo()
					info.Remark = fmt.Sprintf("%s_%s", t.Name(), info.Remark)
					dao := transactionModel.NewDao(ctx.GetDb())
					transaction, err := dao.Create(info, transactionModel.RecordTypeOfManual)
					assert.NoError(t, err)
					assert.NotZero(t, transaction.ID)

					transaction, err = dao.Create(info, transactionModel.RecordTypeOfManual)
					assert.ErrorIs(t, err, global.ErrTransactionSame)
					assert.NotZero(t, transaction.ID)

					hashBytes, err := transaction.Hash()
					assert.NoError(t, err)
					err = tx.Where(
						"account_id = ? AND hash = ?", transaction.AccountId, hashBytes,
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

	// Scenario 2: Update operation when Hash records are missing
	t.Run(
		"UpdateTransactionWithMissingHash", func(t *testing.T) {
			_ = db.Transaction(
				context.TODO(), func(ctx *cus.TxContext) error {
					tx, info := ctx.GetDb(), get.TransInfo()
					dao := transactionModel.NewDao(tx)
					transaction, err := dao.Create(info, transactionModel.RecordTypeOfManual)
					assert.NoError(t, err)
					assert.NotZero(t, transaction.ID)

					err = tx.Where("trans_id = ?", transaction.ID).Delete(&transactionModel.Hash{}).Error
					assert.NoError(t, err)

					transaction.Remark = fmt.Sprintf("%s_%d", t.Name(), transaction.ID)
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

	// Scenario 3: Conflicting Hash updates
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
					_, err = dao.Create(transaction.Info, transactionModel.RecordTypeOfManual)
					assert.NoError(t, err)
					assert.NotZero(t, transaction.ID)

					err = dao.Update(&transaction)
					assert.ErrorIs(t, err, global.ErrTransactionSame)
					return nil
				},
			)
		},
	)
}
