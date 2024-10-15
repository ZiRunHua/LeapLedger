package nats

import (
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
	"context"
)

type PayloadType interface{}

type handler[Data any] func(Data, context.Context) error

var (
	taskManage  = manager.TaskManage
	eventManage = manager.EventManage
	dlqManage   = manager.DlqManage
)

func init() {
	err := db.InitDb.AutoMigrate(&outbox{})
	if err != nil {
		panic(err)
	}
	SubscribeTaskWithPayloadAndProcessInTransaction(
		TaskOutbox, outboxService.getHandleTransaction(outboxTypeTask),
	)
	SubscribeEvent(
		EventOutbox, "outbox", outboxService.getHandleTransaction(outboxTypeEvent),
	)
}
