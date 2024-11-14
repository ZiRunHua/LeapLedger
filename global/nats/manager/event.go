package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ZiRunHua/LeapLedger/util/dataTool"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// EventManager is used to manage events.
// EventManager has a dedicated consumer group that publishes tasks to TaskManage when events are triggered,
// and new consumer groups can be created to Consume events.
// These consumer groups use the same stream.
type EventManager interface {
	Publish(event Event, payload []byte) bool
	Subscribe(event Event, triggerTask Task, fetchTaskData func(eventData []byte) ([]byte, error))
	SubscribeToNewConsumer(event Event, name string, handler MessageHandler)
	updateAllConsumerConfig(func(*jetstream.ConsumerConfig) error, context.Context) error
}

const (
	natsEventName   = "event"
	natsEventPrefix = "event"
)

var (
	natsEventLogPath = filepath.Join(natsLogPath, "event.log")
)

type Event string

func (t Event) subject() string { return fmt.Sprintf("%s.subject_%s", natsEventPrefix, t) }

const EventRetryTriggerTask Event = "retry_trigger_task"

type RetryTriggerTask struct {
	Task Task
	Data []byte
}

type eventManager struct {
	EventManager
	manageInitializers
	eventMsgHandler

	dlqRegisterStream
}

func (em *eventManager) init(js jetstream.JetStream, taskManage *taskManager, logger *zap.Logger) error {
	em.eventMsgHandler.init()
	em.taskManage = taskManage
	streamConfig := jetstream.StreamConfig{
		Name:      natsEventName,
		Subjects:  []string{natsEventPrefix + ".*"},
		Retention: jetstream.InterestPolicy,
		MaxAge:    24 * time.Hour * 7,
	}
	consumerConfig := jetstream.ConsumerConfig{
		Name:          natsEventPrefix + "_consumer",
		Durable:       natsEventPrefix + "_consumer",
		AckPolicy:     jetstream.AckExplicitPolicy,
		BackOff:       backOff,
		MaxDeliver:    len(backOff) + 1,
		MaxAckPending: runtime.GOMAXPROCS(0) * 3,
	}
	err := em.manageInitializers.init(js, streamConfig, consumerConfig, logger)
	if err != nil {
		return err
	}
	return em.manageInitializers.setMainConsumerConsume(context.TODO(), em.msgHandle)
}

func (em *eventManager) Publish(event Event, payload []byte) bool {
	_, err := em.js.PublishAsync(event.subject(), payload)
	if err != nil {
		em.logger.Error("Publish", zap.Error(err))
		return false
	}
	return true
}

// Subscribe sets up the task triggered by an event.
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
	_, err := em.consumerManger.NewConsumer(
		context.TODO(),
		func(config *jetstream.ConsumerConfig) error {
			config.Name, config.Durable, config.FilterSubjects = name, name, []string{event.subject()}
			return nil
		},
		func(_ string, payload []byte) error { return handler(payload) },
	)
	if err != nil {
		panic(err)
	}
}

func (em *eventManager) updateAllConsumerConfig(
	handle func(*jetstream.ConsumerConfig) error, ctx context.Context,
) error {
	return em.consumerManger.UpdateAllConsumerConfig(handle, ctx)
}

func (em *eventManager) getStreamName() string { return natsEventName }

func (em *eventManager) reConsume(ctx context.Context, consumer string, streamSeq uint64) error {
	rawMsg, err := em.stream.GetMsg(ctx, streamSeq)
	if err != nil {
		return err
	}
	return em.consumerManger.ReConsume(ctx, consumer, rawMsg)
}

type eventMsgHandler struct {
	eventToTask   dataTool.Map[Event, dataTool.Map[Task, MessageHandler]]
	msgHandlerMap dataTool.Map[string, MessageHandler]
	msgManger

	taskManage *taskManager
}

func (em *eventMsgHandler) init() {
	em.eventToTask = dataTool.NewSyncMap[Event, dataTool.Map[Task, MessageHandler]]()
	em.msgHandlerMap = dataTool.NewSyncMap[string, MessageHandler]()
}

func (em *eventMsgHandler) getHandler(subject string) (MessageHandler, error) {
	if subject == EventRetryTriggerTask.subject() {
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
