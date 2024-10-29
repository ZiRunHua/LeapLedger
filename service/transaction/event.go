package transactionService

import (
	"context"
	"github.com/ZiRunHua/LeapLedger/global/nats"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
)

func init() {
	// statistic update
	nats.BindTaskToEventAndMakePayload(
		nats.EventTransactionCreate, nats.TaskStatisticUpdate,
		func(eventData transactionModel.Transaction) (transactionModel.StatisticData, error) {
			return eventData.GetStatisticData(true), nil
		},
	)

	nats.SubscribeEvent(
		nats.EventTransactionUpdate, "update_statistic_after_transaction_update",
		func(eventData nats.EventTransactionUpdatePayload, ctx context.Context) error {
			err := GroupApp.Transaction.updateStatistic(eventData.OldTrans.GetStatisticData(false), ctx)
			if err != nil {
				return err
			}
			return GroupApp.Transaction.updateStatistic(eventData.NewTrans.GetStatisticData(true), ctx)
		},
	)

	nats.BindTaskToEventAndMakePayload(
		nats.EventTransactionDelete, nats.TaskStatisticUpdate,
		func(eventData transactionModel.Transaction) (transactionModel.StatisticData, error) {
			return eventData.GetStatisticData(false), nil
		},
	)
}
