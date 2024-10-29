package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/ZiRunHua/LeapLedger/util/dataTool"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// EventManager is used to manage events.
// EventManager has a dedicated consumer group that publishes tasks to TaskManage when events are triggered,
// and new consumer groups can be created to consume events.
// These consumer groups use the same stream.
type EventManager interface {
	Publish(event Event, payload []byte) bool
	Subscribe(event Event, triggerTask Task, fetchTaskData func(eventData []byte) ([]byte, error))
	SubscribeToNewConsumer(event Event, name string, handler MessageHandler)
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
	em.eventMsgHandler.init(logger)
	em.taskManage = taskManage
	streamConfig := jetstream.StreamConfig{
		Name:      natsEventName,
		Subjects:  []string{natsEventPrefix + ".*"},
		Retention: jetstream.InterestPolicy,
		MaxAge:    24 * time.Hour * 7,
	}
	customerConfig := jetstream.ConsumerConfig{
		Name:          natsEventPrefix + "_customer",
		Durable:       natsEventPrefix + "_customer",
		AckPolicy:     jetstream.AckExplicitPolicy,
		BackOff:       backOff,
		MaxDeliver:    len(backOff) + 1,
		MaxAckPending: runtime.GOMAXPROCS(0) * 3,
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
	taskMap, _ := em.eventToTask.LoadOrStore(event, dataTool.NewSyncMap[Task, MessageHandler]())
	taskMap.Store(
		triggerTask, func(payload []byte) error {
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
		},
	)
	em.updateMsgHandlerMap(
		event.subject(), func(payload []byte) error {
			m, exit := em.eventToTask.Load(event)
			if !exit {
				return nil
			}
			m.Range(
				func(_ Task, handler MessageHandler) bool {
					_ = handler(payload)
					return true
				},
			)
			return nil
		},
	)
}
func (em *eventManager) SubscribeToNewConsumer(event Event, name string, handler MessageHandler) {
	em.msgHandlerMap.LoadOrStore(
		event.subject(), func(payload []byte) error {
			// ignore
			return nil
		},
	)
	ctx := context.Background()
	config, err := em.getCustomerConfig(ctx)
	if err != nil {
		panic(err)
	}
	config.Name, config.Durable, config.FilterSubjects = name, name, []string{event.subject()}
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

type eventMsgHandler struct {
	eventToTask   dataTool.Map[Event, dataTool.Map[Task, MessageHandler]]
	msgHandlerMap dataTool.Map[string, MessageHandler]
	msgManger

	logger *zap.Logger

	taskManage *taskManager
}

func (em *eventMsgHandler) init(logger *zap.Logger) {
	em.logger = logger
	em.eventToTask = dataTool.NewSyncMap[Event, dataTool.Map[Task, MessageHandler]]()
	em.msgHandlerMap = dataTool.NewSyncMap[string, MessageHandler]()
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
	handler, exist := em.msgHandlerMap.Load(subject)
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
	em.msgHandlerMap.Store(subject, handler)
}
