package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"sync"
	"time"
)

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

type PayloadType interface{}
type txHandle[Data any] func(Data, context.Context) error

func Subscribe[T PayloadType](task Task, handleTransaction txHandle[T]) {
	handler := func(msg jetstream.Msg) error {
		var data T
		if err := json.Unmarshal(msg.Data(), &data); err != nil {
			return err
		}
		return handleTransaction(data, context.TODO())
	}
	taskManage.subscribe(task, handler)
}

func Publish[T PayloadType](task Task, payload T) (isSuccess bool) {
	str, err := json.Marshal(&payload)
	if err != nil {
		return false
	}
	return taskManage.publish(task, str)
}

const (
	natsTaskName    = "task"
	natsTaskPrefix  = "task"
	natsTaskLogPath = natsLogPath + "task.log"
)

type Task string

func (t Task) subject() string {
	return natsTaskPrefix + ".subject_" + string(t)
}
func (t Task) queue() string {
	return natsTaskPrefix + ".queue_" + string(t)
}

type taskManager struct {
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

func (tm *taskManager) publish(task Task, payload []byte) bool {
	subject := task.subject()
	_, err := tm.js.PublishMsgAsync(
		&nats.Msg{
			Subject: subject,
			Data:    payload,
			Header:  map[string][]string{msgHeaderKeySubject: {subject}},
		},
	)
	if err != nil {
		tm.logger.Error("publish", zap.Error(err))
		return false
	}
	return true
}

type taskMsgHandler struct {
	msgHandlerMap map[string]MessageHandler
	msgManger

	lock   sync.Mutex
	logger *zap.Logger
}

func (tm *taskMsgHandler) subscribe(task Task, handler MessageHandler) {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	if tm.msgHandlerMap == nil {
		tm.msgHandlerMap = make(map[string]MessageHandler)
	}
	tm.msgHandlerMap[task.subject()] = handler
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
