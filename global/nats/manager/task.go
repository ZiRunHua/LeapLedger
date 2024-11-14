package manager

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ZiRunHua/LeapLedger/util/dataTool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// TaskManager is used to manage tasks.
// All tasks will be placed in a consumer group.
type TaskManager interface {
	Publish(task Task, payload []byte) bool
	Subscribe(task Task, handler MessageHandler)
	GetMessageHandler(task Task) (MessageHandler, error)
}

const (
	natsTaskName   = "task"
	natsTaskPrefix = "task"
)

var (
	natsTaskLogPath = filepath.Join(natsLogPath, "task.log")
)

type Task string

func (t Task) subject() string { return natsTaskPrefix + ".subject_" + string(t) }
func (t Task) queue() string   { return natsTaskPrefix + ".queue_" + string(t) }

type taskManager struct {
	TaskManager
	manageInitializers
	taskMsgHandler
}

func (tm *taskManager) init(js jetstream.JetStream, logger *zap.Logger) error {
	tm.taskMsgHandler.init(logger)
	streamConfig := jetstream.StreamConfig{
		Name:      natsTaskName,
		Subjects:  []string{natsTaskPrefix + ".*"},
		Retention: jetstream.InterestPolicy,
		MaxAge:    24 * time.Hour * 7,
	}
	consumerConfig := jetstream.ConsumerConfig{
		Name:          natsTaskPrefix + "_consumer",
		Durable:       natsTaskPrefix + "_consumer",
		AckPolicy:     jetstream.AckExplicitPolicy,
		BackOff:       backOff,
		MaxDeliver:    len(backOff) + 1,
		MaxAckPending: runtime.GOMAXPROCS(0) * 3,
	}

	err := tm.manageInitializers.init(js, streamConfig, consumerConfig)
	if err != nil {
		return err
	}
	_, err = tm.consumer.Consume(tm.receiveMsg)
	return err
}

func (tm *taskManager) Publish(task Task, payload []byte) bool {
	subject := task.subject()
	_, err := tm.js.PublishMsgAsync(
		&nats.Msg{
			Subject: subject,
			Data:    payload,
			Header:  map[string][]string{msgHeaderKeySubject: {subject}},
		},
	)
	if err != nil {
		tm.logger.Error("Publish", zap.Error(err))
		return false
	}
	return true
}

func (tm *taskManager) Subscribe(task Task, handler MessageHandler) {
	tm.msgHandlerMap.Store(task.subject(), handler)
}

func (tm *taskManager) GetMessageHandler(task Task) (MessageHandler, error) {
	return tm.getHandler(task.subject())
}

type taskMsgHandler struct {
	msgHandlerMap dataTool.Map[string, MessageHandler]
	msgManger

	logger *zap.Logger
}

func (tm *taskMsgHandler) init(logger *zap.Logger) {
	tm.logger = logger
	tm.msgHandlerMap = dataTool.NewSyncMap[string, MessageHandler]()
}

func (tm *taskMsgHandler) receiveMsg(msg jetstream.Msg) {
	receiveMsg(msg, func(msg jetstream.Msg) error { return tm.msgHandle(msg.Subject(), msg.Data()) }, tm.logger)
}
func (tm *taskMsgHandler) getHandler(subject string) (MessageHandler, error) {
	handler, exist := tm.msgHandlerMap.Load(subject)
	if !exist {
		return nil, fmt.Errorf("subject: %s ,%w", subject, ErrMsgHandlerNotExist)
	}
	return handler, nil
}

func (tm *taskMsgHandler) msgHandle(subject string, payload []byte) error {
	handler, err := tm.getHandler(subject)
	if err != nil {
		return err
	}
	return handler(payload)
}
