package nats

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ZiRunHua/LeapLedger/global/cus"
	"github.com/ZiRunHua/LeapLedger/global/db"
	"github.com/ZiRunHua/LeapLedger/global/nats/manager"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
)

type Event = manager.Event

const EventOutbox Event = "outbox"
const EventTransactionCreate Event = "transaction_create_event"
const EventTransactionUpdate Event = "transaction_update_event"

type EventTransactionUpdatePayload struct {
	OldTrans, NewTrans transactionModel.Transaction
}

const EventTransactionDelete Event = "transaction_delete_event"

func PublishEvent(event Event) (isSuccess bool) {
	return eventManage.Publish(event, []byte{})
}

func PublishEventWithPayload[T PayloadType](event Event, fetchTaskData T) (isSuccess bool) {
	str, err := json.Marshal(&fetchTaskData)
	if err != nil {
		return false
	}
	return eventManage.Publish(event, str)
}

func SubscribeEvent[T PayloadType](event Event, name string, handleTransaction handler[T]) {
	eventManage.SubscribeToNewConsumer(
		event, name, func(payload []byte) error {
			var data T
			if err := fromJson(payload, &data); err != nil {
				return err
			}
			return db.Transaction(
				context.TODO(), func(ctx *cus.TxContext) error {
					return handleTransaction(data, ctx)
				},
			)
		},
	)
}

func BindTaskToEvent(event Event, triggerTask Task) {
	eventManage.Subscribe(
		event, triggerTask,
		func(eventData []byte) ([]byte, error) {
			return eventData, nil
		},
	)
}

func BindTaskToEventAndMakePayload[T PayloadType, TriggerTaskDataType PayloadType](
	event Event, triggerTask Task, fetchTaskData func(eventData T) (TriggerTaskDataType, error),
) {
	eventManage.Subscribe(
		event, triggerTask, func(eventData []byte) ([]byte, error) {
			var data T
			if err := fromJson(eventData, &data); err != nil {
				return nil, err
			}
			taskData, err := fetchTaskData(data)
			if err != nil {
				return nil, err
			}
			return json.Marshal(taskData)
		},
	)
}

func PublishEventToOutboxWithPayload[T PayloadType](ctx context.Context, event Event, payload T) error {
	if event == EventOutbox {
		return errors.New("cannot be TaskOutbox")
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	id, err := outboxService.sendToOutbox(db.Get(ctx), outboxTypeEvent, string(event), bytes)
	if err != nil {
		return err
	}
	return db.AddCommitCallback(ctx, func() { PublishEventWithPayload(EventOutbox, id) })
}
