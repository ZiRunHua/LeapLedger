package nats

import (
	"context"
	"encoding/json"
	"github.com/ZiRunHua/LeapLedger/global/db"
	"github.com/ZiRunHua/LeapLedger/global/nats/manager"
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
	SubscribeTaskWithPayload(
		TaskOutbox, outboxService.getHandleTransaction(outboxTypeTask),
	)
	SubscribeEvent(
		EventOutbox, "outbox", outboxService.getHandleTransaction(outboxTypeEvent),
	)
}

func fromJson[T PayloadType](jsonStr []byte, data *T) error {
	if len(jsonStr) != 0 {
		if err := json.Unmarshal(jsonStr, &data); err != nil {
			return err
		}
	}
	return nil
}
