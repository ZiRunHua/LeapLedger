package nats

import (
	"context"
	"encoding/json"
	"errors"

	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
)

type Task = manager.Task

const TaskOutbox Task = "outbox"

// email task
const TaskSendCaptchaEmail Task = "sendCaptchaEmail"

type PayloadSendCaptchaEmail struct {
	Email  string
	Action constant.UserAction
}

const TaskSendNotificationEmail Task = "sendNotificationEmail"

type PayloadSendNotificationEmail struct {
	UserId       uint
	Notification constant.Notification
}

// user task
const TaskCreateTourist Task = "createTourist"

// transaction task
const TaskStatisticUpdate Task = "statisticUpdate"
const TaskTransactionSync Task = "transactionSync"
const TaskTransactionTimingExec Task = "transactionTimingExec"
const TaskTransactionTimingTaskAssign Task = "transactionTimingTaskAssign"

// category task
const TaskMappingCategoryToAccountMapping Task = "mappingCategoryToAccountMapping"
const TaskUpdateCategoryMapping Task = "updateCategoryMapping"

func PublishTask(task Task) (isSuccess bool) {
	return taskManage.Publish(task, []byte{})
}

func SubscribeTask(task Task, handler func(ctx context.Context) error) {
	taskManage.Subscribe(task, func(payload []byte) error { return handler(context.Background()) })
}

func PublishTaskWithPayload[T PayloadType](task Task, payload T) (isSuccess bool) {
	str, err := json.Marshal(&payload)
	if err != nil {
		return false
	}
	return taskManage.Publish(task, str)
}

func SubscribeTaskWithPayload[T PayloadType](task Task, handle handler[T]) {
	msgHandler := func(payload []byte) error {
		var data T
		if err := json.Unmarshal(payload, &data); err != nil {
			return err
		}
		return handle(data, context.Background())
	}
	taskManage.Subscribe(task, msgHandler)
}

func SubscribeTaskWithPayloadAndProcessInTransaction[T PayloadType](task Task, handleTransaction handler[T]) {
	taskManage.Subscribe(
		task, func(payload []byte) error {
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

// PublishTaskToOutbox
// Must be used in a transaction,see db.Transaction
func PublishTaskToOutbox(ctx context.Context, task Task) error {
	if task == TaskOutbox {
		return errors.New("task cannot be TaskOutbox")
	}
	id, err := outboxService.sendToOutbox(db.Get(ctx), outboxTypeTask, string(task), []byte{})
	if err != nil {
		return err
	}
	return db.AddCommitCallback(ctx, func() { PublishTaskWithPayload(TaskOutbox, id) })
}

// PublishTaskToOutboxWithPayload
// Must be used in a transaction,see db.Transaction
func PublishTaskToOutboxWithPayload[T PayloadType](ctx context.Context, task Task, payload T) error {
	if task == TaskOutbox {
		return errors.New("task cannot be TaskOutbox")
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	id, err := outboxService.sendToOutbox(db.Get(ctx), outboxTypeTask, string(task), bytes)
	if err != nil {
		return err
	}
	return db.AddCommitCallback(ctx, func() { PublishTaskWithPayload(TaskOutbox, id) })
}
