package nats

import (
	"KeepAccount/global/nats/manager"
	"context"
)

type PayloadType interface{}

type handle[Data any] func(Data, context.Context) error

var (
	taskManage  = manager.TaskManage
	eventManage = manager.EventManage
	dlqManage   = manager.DlqManage
)
