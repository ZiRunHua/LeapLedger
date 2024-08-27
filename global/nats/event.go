package nats

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

func PublishEvent(event Event) (isSuccess bool) {
	return eventManage.publish(event, []byte{})
}

func SubscribeEvent(event Event, triggerTask Task) {
	eventManage.subscribe(event, triggerTask, func(eventData []byte) ([]byte, error) { return []byte{}, nil })
}

func PublishEventWithPayload[EventDataType PayloadType](event Event, fetchTaskData EventDataType) (isSuccess bool) {
	str, err := json.Marshal(&fetchTaskData)
	if err != nil {
		return false
	}
	return eventManage.publish(event, str)
}

func SubscribeEventWithPayload[EventDataType PayloadType, TriggerTaskDataType PayloadType](
	event Event, triggerTask Task, fetchTaskData func(eventData EventDataType) (TriggerTaskDataType, error),
) {
	eventManage.subscribe(
		event, triggerTask, func(eventData []byte) ([]byte, error) {
			var data EventDataType
			if err := json.Unmarshal(eventData, &data); err != nil {
				return nil, err
			}
			taskData, err := fetchTaskData(data)
			if err != nil {
				return nil, err
			}
			return json.Marshal(taskData)
		},
	)
}

const (
	NatsEventName    = "event"
	NatsEventPrefix  = "event"
	NatsEventLogPath = natsLogPath + "event.log"
)

type Event string

func (t Event) subject() string {
	return fmt.Sprintf("%s.subject_%s", NatsEventPrefix, t)
}

func (t Event) queue() string {
	return fmt.Sprintf("%s.queue_%s", NatsEventPrefix, t)
}

const EventRetryTriggerTask Event = "retry_trigger_task"

type RetryTriggerTask struct {
	Task Task
	Data []byte
}

type eventManager struct {
	manageInitializers
	eventMsgHandler
}

func (em *eventManager) init(js jetstream.JetStream, taskManage *taskManager, logger *zap.Logger) error {
	em.taskManage, em.logger = taskManage, logger
	streamConfig := jetstream.StreamConfig{
		Name:      NatsEventName,
		Subjects:  []string{NatsEventPrefix + ".*"},
		Retention: jetstream.InterestPolicy,
		MaxAge:    24 * time.Hour * 7,
	}
	customerConfig := jetstream.ConsumerConfig{
		Name:       NatsEventPrefix + "_customer",
		Durable:    NatsEventPrefix + "_customer",
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

func (em *eventManager) publish(event MsgType, payload []byte) bool {
	_, err := em.js.PublishAsync(event.subject(), payload)
	if err != nil {
		em.logger.Error("publish", zap.Error(err))
		return false
	}
	return true
}

func (em *eventManager) subscribe(event Event, triggerTask Task, fetchTaskData func(eventData []byte) ([]byte, error)) {
	em.lock.Lock()
	defer em.lock.Unlock()
	if em.eventToTask == nil {
		em.eventToTask = make(map[Event]map[Task]MessageHandler)
		if em.eventToTask[event] == nil {
			em.eventToTask[event] = make(map[Task]MessageHandler)
		}
	}
	em.eventToTask[event][triggerTask] = func(msg jetstream.Msg) error {
		data, err := fetchTaskData(msg.Data())
		if err != nil {
			return err
		}
		if em.taskManage.publish(triggerTask, data) {
			return nil
		}
		str, _ := json.Marshal(RetryTriggerTask{Task: triggerTask, Data: msg.Data()})
		em.publish(EventRetryTriggerTask, str)
		return nil
	}

	if em.msgHandlerMap == nil {
		em.msgHandlerMap = make(map[string]MessageHandler)
	}
	if em.msgHandlerMap[event.subject()] != nil {
		return
	}
	em.msgHandlerMap[event.subject()] = func(msg jetstream.Msg) error {
		for _, handler := range em.eventToTask[event] {
			_ = handler(msg)
		}
		return nil
	}
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
	receiveMsg(msg, func(msg jetstream.Msg) error { return em.msgHandle(msg) }, em.logger)
}

func (em *eventMsgHandler) getHandler(subject string) (MessageHandler, error) {
	if subject == string(EventRetryTriggerTask) {
		return func(msg jetstream.Msg) error {
			var data RetryTriggerTask
			err := json.Unmarshal(msg.Data(), &data)
			if err != nil {
				return err
			}
			isSuccess := em.taskManage.publish(data.Task, data.Data)
			if !isSuccess {
				return errors.New("retry event trigger task fail")
			}
			return nil
		}, nil
	}
	handler, exist := em.msgHandlerMap[subject]
	if !exist {
		return nil, fmt.Errorf("subject: %s ,%w", subject, ErrMsgHandlerNotExist)
	}
	return handler, nil
}

func (em *eventMsgHandler) msgHandle(msg jetstream.Msg) error {
	handler, err := em.getHandler(msg.Subject())
	if err != nil {
		return err
	}
	return handler(msg)
}
