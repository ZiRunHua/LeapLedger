package transactionService

import (
	"KeepAccount/global/contextKey"
	gTask "KeepAccount/global/task"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"gorm.io/gorm"
)

type _task struct{}

func init() {
	gTask.Subscribe[transactionModel.StatisticData](
		gTask.TaskStatisticUpdate, func(data transactionModel.StatisticData, ctx context.Context) error {
			return GroupApp.Transaction.updateStatistic(data, ctx.Value(contextKey.Tx).(*gorm.DB))
		},
	)

	gTask.Subscribe[transactionModel.Transaction](
		gTask.TaskTransactionSync, GroupApp.Transaction.SyncToMappingAccount,
	)
}

func (t *_task) updateStatistic(data transactionModel.StatisticData, tx *gorm.DB) error {
	isSuccess := gTask.Publish[transactionModel.StatisticData](gTask.TaskStatisticUpdate, data)
	if false == isSuccess {
		return server.updateStatistic(data, tx)
	}
	return nil
}

func (t *_task) syncToMappingAccount(trans transactionModel.Transaction, ctx context.Context) error {
	isSuccess := gTask.Publish[transactionModel.Transaction](gTask.TaskTransactionSync, trans)
	if false == isSuccess {
		return server.SyncToMappingAccount(trans, ctx)
	}
	return nil
}
