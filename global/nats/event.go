package nats

import (
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go/jetstream"
)

type Event manager.Event

const EventTransactionCreate Event = "transaction_create_event"
const EventTransactionUpdate Event = "transaction_update_event"

type EventTransactionUpdatePayload struct {
	OldTrans, NewTrans transactionModel.Transaction
}

const EventTransactionDelete Event = "transaction_delete_event"

func PublishEvent(event Event) (isSuccess bool) {
	return eventManage.Publish(manager.Event(event), []byte{})
}

func PublishEventWithPayload[EventDataType PayloadType](event Event, fetchTaskData EventDataType) (isSuccess bool) {
	str, err := json.Marshal(&fetchTaskData)
	if err != nil {
		return false
	}
	return eventManage.Publish(manager.Event(event), str)
}

func SubscribeEvent[EventDataType PayloadType](event Event, name string, handleTransaction handle[EventDataType]) {
	handler := func(msg jetstream.Msg) error {
		var data EventDataType
		if err := json.Unmarshal(msg.Data(), &data); err != nil {
			return err
		}
		return db.Transaction(context.TODO(), func(ctx *cus.TxContext) error {
			return handleTransaction(data, ctx)
		})
	}
	eventManage.SubscribeToNewConsumer(manager.Event(event), name, handler)
}

func BindTaskToEvent(event Event, triggerTask Task) {
	eventManage.Subscribe(manager.Event(event), manager.Task(triggerTask), func(eventData []byte) ([]byte, error) { return []byte{}, nil })
}

func BindTaskToEventAndMakePayload[EventDataType PayloadType, TriggerTaskDataType PayloadType](
	event Event, triggerTask Task, fetchTaskData func(eventData EventDataType) (TriggerTaskDataType, error),
) {
	eventManage.Subscribe(
		manager.Event(event), manager.Task(triggerTask), func(eventData []byte) ([]byte, error) {
			var data EventDataType
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
