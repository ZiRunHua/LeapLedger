package manager

import (
	"runtime/debug"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/initialize"
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

	TaskManage  TaskManager
	EventManage EventManager
	DlqManage   DlqManager
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
	taskLogger, err = global.Config.Logger.New(natsTaskLogPath)
	if err != nil {
		panic(err)
	}
	eventLogger, err = global.Config.Logger.New(natsEventLogPath)
	if err != nil {
		panic(err)
	}
	dlqLogger, err = global.Config.Logger.New(dlqLogPath)
	if err != nil {
		panic(err)
	}

	if taskManage != nil {
		taskManage = &taskManager{taskMsgHandler: taskManage.taskMsgHandler}
	} else {
		taskManage = &taskManager{}
	}
	TaskManage = taskManage
	err = taskManage.init(js, taskLogger)
	if err != nil {
		panic(err)
	}

	if eventManage != nil {
		eventManage = &eventManager{eventMsgHandler: eventManage.eventMsgHandler}
	} else {
		eventManage = &eventManager{}
	}
	EventManage = eventManage
	err = eventManage.init(js, taskManage, eventLogger)
	if err != nil {
		panic(err)
	}

	dlqManage = &dlqManager{}
	DlqManage = dlqManage
	err = dlqManage.init(js, []jetstream.Stream{taskManage.stream, eventManage.stream}, dlqLogger)
	if err != nil {
		panic(err)
	}
}

func receiveMsg(msg jetstream.Msg, handle func(msg jetstream.Msg) error, logger *zap.Logger) {
	var err error
	defer func() {
		r := recover()
		if r == nil {
			if err != nil {
				logger.Error("receiveMsg err", zap.Error(err))
				err = msg.Nak()
			} else {
				err = msg.Ack()
			}
		} else {
			logger.Error("receiveMsg panic", zap.Any("panic", r), zap.Stack(string(debug.Stack())))
			err = msg.Nak()
		}
		if err != nil {
			logger.Error("receiveMsg ack err", zap.Error(err))
		}
	}()
	err = handle(msg)
}
