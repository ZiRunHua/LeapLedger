package transactionService

import (
	"KeepAccount/global/contextKey"
	"KeepAccount/global/nats"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"gorm.io/gorm"
)

type _task struct{}

func init() {
	nats.TransSubscribe[transactionModel.StatisticData](
		nats.TaskStatisticUpdate, func(db *gorm.DB, data transactionModel.StatisticData) error {
			return GroupApp.Transaction.updateStatistic(data, db)
		},
	)

	nats.TransSubscribe[transactionModel.Transaction](
		nats.TaskTransactionSync, func(db *gorm.DB, data transactionModel.Transaction) error {
			return GroupApp.Transaction.SyncToMappingAccount(data, context.WithValue(context.Background(), contextKey.Tx, db))
		},
	)
}

func (t *_task) updateStatistic(data transactionModel.StatisticData, tx *gorm.DB) error {
	isSuccess := nats.Publish[transactionModel.StatisticData](nats.TaskStatisticUpdate, data)
	if false == isSuccess {
		return server.updateStatistic(data, tx)
	}
	return nil
}

func (t *_task) syncToMappingAccount(trans transactionModel.Transaction, ctx context.Context) error {
	isSuccess := nats.Publish[transactionModel.Transaction](nats.TaskTransactionSync, trans)
	if false == isSuccess {
		return server.SyncToMappingAccount(trans, ctx)
	}
	return nil
}
