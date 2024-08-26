package nats

import (
	"KeepAccount/global/constant"
	"KeepAccount/initialize"
	"KeepAccount/util/log"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

var (
	natsConn = initialize.Nats
	js       jetstream.JetStream
)

var (
	taskManage  *taskManager
	eventManage *eventManager
	dlqManage   *dlqManager
)

const natsLogPath = constant.LOG_PATH + "/nats/"

var (
	taskLogger  *zap.Logger
	eventLogger *zap.Logger
	dlqLogger   *zap.Logger
)

func init() {
	var err error
	js, err = jetstream.New(natsConn)
	if err != nil {
		panic(err)
	}
	taskLogger, err = log.GetNewZapLogger(natsTaskLogPath)
	if err != nil {
		panic(err)
	}
	eventLogger, err = log.GetNewZapLogger(NatsEventLogPath)
	if err != nil {
		panic(err)
	}
	dlqLogger, err = log.GetNewZapLogger(dlqLogPath)
	if err != nil {
		panic(err)
	}

	taskManage = &taskManager{}
	err = taskManage.init(js, taskLogger)
	if err != nil {
		panic(err)
	}
	eventManage = &eventManager{}
	err = eventManage.init(js, taskManage, eventLogger)
	if err != nil {
		panic(err)
	}
	dlqManage = &dlqManager{}
	err = dlqManage.init(js, []jetstream.Stream{taskManage.stream, eventManage.stream}, dlqLogger)
	if err != nil {
		panic(err)
	}
}

func receiveMsg(msg jetstream.Msg, handle MessageHandler, logger *zap.Logger) {
	var err error
	defer func() {
		r := recover()
		if r == nil {
			if err != nil {
				err = msg.Nak()
			} else {
				err = msg.Ack()
			}
		} else {
			logger.Error("receiveMsg", zap.Any("panic", r))
			err = msg.Nak()
		}
		if err != nil {
			logger.Error("receiveMsg", zap.Error(err))
		}
	}()
	err = handle(msg)
}
