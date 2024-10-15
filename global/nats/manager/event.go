package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type EventManager interface {
	Publish(event Event, payload []byte) bool
	Subscribe(event Event, triggerTask Task, fetchTaskData func(eventData []byte) ([]byte, error))
	SubscribeToNewConsumer(event Event, name string, handler MessageHandler)
	GetMessageHandler(event Event) (MessageHandler, error)
}

const (
	natsEventName    = "event"
	natsEventPrefix  = "event"
	natsEventLogPath = natsLogPath + "event.log"
)

type Event string

func (t Event) subject() string { return fmt.Sprintf("%s.subject_%s", natsEventPrefix, t) }

func (t Event) queue() string { return fmt.Sprintf("%s.queue_%s", natsEventPrefix, t) }

const EventRetryTriggerTask Event = "retry_trigger_task"

type RetryTriggerTask struct {
	Task Task
	Data []byte
}

type eventManager struct {
	EventManager
	manageInitializers
	eventMsgHandler
}

func (em *eventManager) init(js jetstream.JetStream, taskManage *taskManager, logger *zap.Logger) error {
	em.taskManage, em.logger = taskManage, logger
	streamConfig := jetstream.StreamConfig{
		Name:      natsEventName,
		Subjects:  []string{natsEventPrefix + ".*"},
		Retention: jetstream.InterestPolicy,
		MaxAge:    24 * time.Hour * 7,
	}
	customerConfig := jetstream.ConsumerConfig{
		Name:       natsEventPrefix + "_customer",
		Durable:    natsEventPrefix + "_customer",
		AckPolicy:  jetstream.AckExplicitPolicy,
		BackOff:    backOff,
		MaxDeliver: len(backOff) + 1,
	}
	err := em.manageInitializers.init(js, streamConfig, customerConfig)
	if err != nil {
		return err
	}
	_, err = em.consumer.Consume(em.receiveMsg)
	return err
}

func (em *eventManager) Publish(event Event, payload []byte) bool {
	_, err := em.js.PublishAsync(event.subject(), payload)
	if err != nil {
		em.logger.Error("Publish", zap.Error(err))
		return false
	}
	return true
}

func (em *eventManager) Subscribe(event Event, triggerTask Task, fetchTaskData func(eventData []byte) ([]byte, error)) {
	em.lock.Lock()
	defer em.lock.Unlock()
	if em.eventToTask == nil {
		em.eventToTask = make(map[Event]map[Task]MessageHandler)
	}
	if em.eventToTask[event] == nil {
		em.eventToTask[event] = make(map[Task]MessageHandler)
	}
	em.eventToTask[event][triggerTask] = func(payload []byte) error {
		data, err := fetchTaskData(payload)
		if err != nil {
			return err
		}
		if em.taskManage.Publish(triggerTask, data) {
			return nil
		}
		str, _ := json.Marshal(RetryTriggerTask{Task: triggerTask, Data: payload})
		em.Publish(EventRetryTriggerTask, str)
		return nil
	}

	em.updateMsgHandlerMap(
		event.subject(), func(payload []byte) error {
			for _, handler := range em.eventToTask[event] {
				_ = handler(payload)
			}
			return nil
		},
	)
}
func (em *eventManager) SubscribeToNewConsumer(event Event, name string, handler MessageHandler) {
	updateMsg := func() {
		em.lock.Lock()
		defer em.lock.Unlock()
		if _, exist := em.msgHandlerMap[event.subject()]; exist {
			return
		}
		em.updateMsgHandlerMap(
			event.subject(), func(payload []byte) error {
				// ignore
				return nil
			},
		)
	}
	updateMsg()
	ctx := context.Background()
	config, err := em.getCustomerConfig(ctx)
	if err != nil {
		panic(err)
	}
	config.Name, config.Durable, config.FilterSubjects = name, name, []string{string(event.subject())}
	customer, err := em.newCustomer(ctx, config)
	if err != nil {
		panic(err)
	}
	_, err = customer.Consume(
		func(msg jetstream.Msg) {
			receiveMsg(msg, func(msg jetstream.Msg) error { return handler(msg.Data()) }, em.logger)
		},
	)
	if err != nil {
		panic(err)
	}
}

func (em *eventManager) GetMessageHandler(event Event) (MessageHandler, error) {
	return em.getHandler(event.subject())
}

type eventMsgHandler struct {
	eventToTask   map[Event]map[Task]MessageHandler
	msgHandlerMap map[string]MessageHandler
	msgManger

	lock   sync.Mutex
	logger *zap.Logger

	taskManage *taskManager
}

func (em *eventMsgHandler) receiveMsg(msg jetstream.Msg) {
	receiveMsg(msg, func(msg jetstream.Msg) error { return em.msgHandle(msg.Subject(), msg.Data()) }, em.logger)
}

func (em *eventMsgHandler) getHandler(subject string) (MessageHandler, error) {
	if subject == string(EventRetryTriggerTask) {
		return func(payload []byte) error {
			var data RetryTriggerTask
			err := json.Unmarshal(payload, &data)
			if err != nil {
				return err
			}
			isSuccess := em.taskManage.Publish(data.Task, data.Data)
			if !isSuccess {
				return errors.New("retry event trigger task fail")
			}
			return nil
		}, nil
	}
	handler, exist := em.msgHandlerMap[subject]
	if !exist {
		return func(payload []byte) error { return nil }, fmt.Errorf(
			"subject: %s ,%w", subject, ErrMsgHandlerNotExist,
		)
	}
	return handler, nil
}

func (em *eventMsgHandler) msgHandle(subject string, payload []byte) error {
	handler, err := em.getHandler(subject)
	if err != nil {
		return err
	}
	return handler(payload)
}

func (em *eventMsgHandler) updateMsgHandlerMap(subject string, handler MessageHandler) {
	if em.msgHandlerMap == nil {
		em.msgHandlerMap = make(map[string]MessageHandler)
	}
	em.msgHandlerMap[subject] = handler
}
