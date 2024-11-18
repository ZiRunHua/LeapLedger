package manager

import (
	"context"
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

	dlqRegisterStream
}

func (tm *taskManager) init(js jetstream.JetStream, logger *zap.Logger) error {
	tm.taskMsgHandler.init()
	streamConfig := jetstream.StreamConfig{
		Name:      natsTaskName,
		Subjects:  []string{natsTaskPrefix + ".*"},
		Retention: jetstream.LimitsPolicy,
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

	err := tm.manageInitializers.init(js, streamConfig, consumerConfig, logger)
	if err != nil {
		return err
	}
	return tm.manageInitializers.setMainConsumerConsume(context.TODO(), tm.msgHandle)
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

func (tm *taskManager) getStreamName() string { return natsTaskName }

func (tm *taskManager) reConsume(ctx context.Context, _ string, streamSeq uint64) error {
	rawMsg, err := tm.stream.GetMsg(ctx, streamSeq)

	if err != nil {
		return err
	}
	// taskManager has only one consumer group, so it uses the PublishMsgAsync method directly
	_, err = tm.js.PublishMsgAsync(
		&nats.Msg{
			Subject: rawMsg.Subject,
			Data:    rawMsg.Data,
			Header:  rawMsg.Header,
		},
	)
	return err
}

type taskMsgHandler struct {
	msgHandlerMap dataTool.Map[string, MessageHandler]
	msgManger
}

func (tm *taskMsgHandler) init() {
	tm.msgHandlerMap = dataTool.NewSyncMap[string, MessageHandler]()
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
