package nats

import (
	"KeepAccount/global/cusCtx"
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go/jetstream"
)

type Task manager.Task

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
	return taskManage.Publish(manager.Task(task), []byte{})
}

func SubscribeTask(task Task, handler func() error) {
	taskManage.Subscribe(manager.Task(task), func(msg jetstream.Msg) error { return handler() })
}

func PublishTaskWithPayload[T PayloadType](task Task, payload T) (isSuccess bool) {
	str, err := json.Marshal(&payload)
	if err != nil {
		return false
	}
	return taskManage.Publish(manager.Task(task), str)
}

func SubscribeTaskWithPayload[T PayloadType](task Task, handleTransaction txHandle[T]) {
	handler := func(msg jetstream.Msg) error {
		var data T
		if err := json.Unmarshal(msg.Data(), &data); err != nil {
			return err
		}
		return db.Transaction(context.TODO(), func(ctx *cusCtx.TxContext) error {
			return handleTransaction(data, ctx)
		})
	}
	taskManage.Subscribe(manager.Task(task), handler)
}
