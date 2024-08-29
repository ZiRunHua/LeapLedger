package manager

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"sync"
	"time"
)

type TaskManager interface {
	Publish(task Task, payload []byte) bool
	Subscribe(task Task, handler MessageHandler)
}

const (
	natsTaskName    = "task"
	natsTaskPrefix  = "task"
	natsTaskLogPath = natsLogPath + "task.log"
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
	tm.logger = logger
	streamConfig := jetstream.StreamConfig{
		Name:      natsTaskName,
		Subjects:  []string{natsTaskPrefix + ".*"},
		Retention: jetstream.InterestPolicy,
		MaxAge:    24 * time.Hour * 7,
	}
	customerConfig := jetstream.ConsumerConfig{
		Name:       natsTaskPrefix + "_customer",
		Durable:    natsTaskPrefix + "_customer",
		AckPolicy:  jetstream.AckExplicitPolicy,
		BackOff:    backOff,
		MaxDeliver: len(backOff) + 1,
	}
	err := tm.manageInitializers.init(js, streamConfig, customerConfig)
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
	tm.lock.Lock()
	defer tm.lock.Unlock()
	if tm.msgHandlerMap == nil {
		tm.msgHandlerMap = make(map[string]MessageHandler)
	}
	tm.msgHandlerMap[task.subject()] = handler
}

type taskMsgHandler struct {
	msgHandlerMap map[string]MessageHandler
	msgManger

	lock   sync.Mutex
	logger *zap.Logger
}

func (tm *taskMsgHandler) receiveMsg(msg jetstream.Msg) {
	receiveMsg(msg, func(msg jetstream.Msg) error { return tm.msgHandle(msg) }, tm.logger)
}
func (tm *taskMsgHandler) getHandler(subject string) (MessageHandler, error) {
	handler, exist := tm.msgHandlerMap[subject]
	if !exist {
		return nil, fmt.Errorf("subject: %s ,%w", subject, ErrMsgHandlerNotExist)
	}
	return handler, nil
}

func (tm *taskMsgHandler) msgHandle(msg jetstream.Msg) error {
	handler, err := tm.getHandler(msg.Subject())
	if err != nil {
		return err
	}
	return handler(msg)
}
