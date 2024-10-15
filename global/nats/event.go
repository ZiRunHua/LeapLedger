package nats

import (
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"encoding/json"
	"errors"
)

type Event manager.Event

const EventOutbox Event = "outbox"
const EventTransactionCreate Event = "transaction_create_event"
const EventTransactionUpdate Event = "transaction_update_event"

type EventTransactionUpdatePayload struct {
	OldTrans, NewTrans transactionModel.Transaction
}

const EventTransactionDelete Event = "transaction_delete_event"

func PublishEvent(event Event) (isSuccess bool) {
	return eventManage.Publish(manager.Event(event), []byte{})
}

func PublishEventWithPayload[T PayloadType](event Event, fetchTaskData T) (isSuccess bool) {
	str, err := json.Marshal(&fetchTaskData)
	if err != nil {
		return false
	}
	return eventManage.Publish(manager.Event(event), str)
}

func SubscribeEvent[T PayloadType](event Event, name string, handleTransaction handler[T]) {
	eventManage.SubscribeToNewConsumer(
		manager.Event(event), name, func(payload []byte) error {
			var data T
			if err := json.Unmarshal(payload, &data); err != nil {
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
		manager.Event(event), manager.Task(triggerTask),
		func(eventData []byte) ([]byte, error) {
			return eventData, nil
		},
	)
}

func BindTaskToEventAndMakePayload[T PayloadType, TriggerTaskDataType PayloadType](
	event Event, triggerTask Task, fetchTaskData func(eventData T) (TriggerTaskDataType, error),
) {
	eventManage.Subscribe(
		manager.Event(event), manager.Task(triggerTask), func(eventData []byte) ([]byte, error) {
			var data T
			if err := json.Unmarshal(eventData, &data); err != nil {
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
