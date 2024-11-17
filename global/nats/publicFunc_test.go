package nats

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ZiRunHua/LeapLedger/global/cus"
	"github.com/ZiRunHua/LeapLedger/global/db"
	"github.com/ZiRunHua/LeapLedger/global/nats/manager"
	"github.com/google/uuid"
)

func init() {
	manager.UpdateTestBackOff()
}

type taskData struct {
	Time int64
	Name string
}

func newTaskData() taskData {
	return taskData{
		Time: time.Now().Unix(),
		Name: uuid.NewString(),
	}
}
func TestTaskPublishAndSubscribe(t *testing.T) {
	task := Task(t.Name() + uuid.NewString())
	msgChan := make(chan struct{})
	SubscribeTask(
		task, func(ctx context.Context) error {
			msgChan <- struct{}{}
			return nil
		},
	)
	t.Run(
		"Publish task", func(t *testing.T) {
			t.Parallel()
			if !PublishTask(task) {
				t.Error("Publish fail")
			}
			<-msgChan
			close(msgChan)
		},
	)
	withPayloadTask, data := Task(t.Name()+uuid.NewString()), newTaskData()
	msgWithDataChan := make(chan struct{ Data taskData })
	SubscribeTaskWithPayloadAndProcessInTransaction(
		withPayloadTask, func(pushData taskData, ctx context.Context) error {
			msgWithDataChan <- struct{ Data taskData }{Data: pushData}
			return nil
		},
	)
	t.Run(
		"Publish task With payload", func(t *testing.T) {
			t.Parallel()
			if !PublishTaskWithPayload(withPayloadTask, data) {
				t.Error("Publish fail")
			}
			msg := <-msgWithDataChan
			if !reflect.DeepEqual(msg.Data, data) {
				t.Fatal("push data not equal:", msg.Data, data)
			}
			close(msgWithDataChan)
		},
	)
}

func TestEventPublishAndSubscribe(t *testing.T) {
	var taskCount = 10
	taskMap, event := make(map[Task]int), Event(uuid.NewString())
	taskChan := make(chan struct{ Task Task }, taskCount)
	for i := 0; i < taskCount; i++ {
		task := Task("task_" + uuid.NewString())
		SubscribeTask(
			task, func(ctx context.Context) error {
				taskChan <- struct{ Task Task }{task}
				return nil
			},
		)
		taskMap[task]++
		BindTaskToEvent(event, task)
	}
	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			PublishEvent(event)
			for true {
				msg, open := <-taskChan
				if !open {
					t.Fatal(errors.New("chan close"))
				}
				taskMap[msg.Task]--
				if taskMap[msg.Task] == 0 {
					delete(taskMap, msg.Task)
				} else if taskMap[msg.Task] < 0 {
					t.Fatal(msg, errors.New("msg repeat"))
				}
				if len(taskMap) != 0 {
					continue
				}
				if len(taskChan) != 0 {
					t.Fatal(taskChan, errors.New("msg repeat"))
				}
				close(taskChan)
				return
			}

		},
	)
}

func TestSubscribeEvent(t *testing.T) {
	name := t.Name() + uuid.NewString()
	var retryCount atomic.Int32
	eventToTask := map[Event]map[Task]struct{}{
		Event(name + "_event_1"): {
			Task(name + "_task_1"): struct{}{},
			Task(name + "_task_2"): struct{}{},
			Task(name + "_task_3"): struct{}{},
		},
	}
	eventSubmitCount := make(map[Event]int)
	msgChan := make(chan struct{ Event Event })
	// subscribe
	for event, tasks := range eventToTask {
		eventSubmitCount[event] = 1
		SubscribeEvent(
			event, string(event)+"_new_customer_group", func(v int, ctx context.Context) error {
				if retryCount.Add(1) < 10 {
					return errors.New("test retry")
				}
				msgChan <- struct{ Event Event }{event}
				return nil
			},
		)
		for task := range tasks {
			eventSubmitCount[event]++
			BindTaskToEvent(event, task)
			SubscribeTaskWithPayload(
				task, func(v int, ctx context.Context) error {
					if retryCount.Add(1) < 10 {
						return errors.New("test retry")
					}
					msgChan <- struct{ Event Event }{event}
					return nil
				},
			)
		}
	}
	for event := range eventToTask {
		PublishEventWithPayload(event, 1)
	}
	for true {
		msg, open := <-msgChan
		if !open {
			t.Fatal(errors.New("chan close"))
		}
		eventSubmitCount[msg.Event]--
		if eventSubmitCount[msg.Event] == 0 {
			delete(eventSubmitCount, msg.Event)
		} else if eventSubmitCount[msg.Event] < 0 {
			t.Fatal(msg, errors.New("msg repeat"))
		}
		if len(eventSubmitCount) == 0 {
			close(msgChan)
			t.Log("finish")
			return
		}
	}
}

func TestOutboxTask(t *testing.T) {
	var retryCount, taskNumber = 2, 3
	msgChan, taskRetryCount := make(chan struct{}, taskNumber), make(map[Task]*atomic.Int32)
	for i := 0; i < taskNumber; i++ {
		task := Task(t.Name() + "task_" + uuid.NewString())
		taskRetryCount[task] = new(atomic.Int32)
		taskRetryCount[task].Add(int32(retryCount))
		SubscribeTaskWithPayload(
			task, func(data int32, ctx context.Context) error {
				if taskRetryCount[task].Add(-1) >= 0 {
					return errors.New("test retry")
				}
				msgChan <- struct{}{}
				return nil
			},
		)
	}
	for task := range taskRetryCount {
		err := db.Transaction(
			context.TODO(), func(ctx *cus.TxContext) error {
				return PublishTaskToOutboxWithPayload(ctx, task, 1)
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	for true {
		_, open := <-msgChan
		if open {
			taskNumber--
			if taskNumber == 0 {
				close(msgChan)
				return
			}
		} else {
			t.Fatal(errors.New("chan close"))
		}
	}
}

func TestOutboxEvent(t *testing.T) {
	const eventNumber, taskNumber = 10, 3
	eventToTask, eventChan := make(map[Event]map[Task]struct{}), make(
		chan struct {
			Event Event
			Task  Task
		}, eventNumber*taskNumber,
	)
	for i := 0; i < eventNumber; i++ {
		event := Event("event_" + uuid.NewString())
		eventToTask[event] = make(map[Task]struct{})
		for j := 0; j < taskNumber; j++ {
			task := Task("task_" + uuid.NewString())
			eventToTask[event][task] = struct{}{}
			SubscribeTaskWithPayload(
				task, func(t int, ctx context.Context) error {
					eventChan <- struct {
						Event Event
						Task  Task
					}{Event: event, Task: task}
					return nil
				},
			)
			BindTaskToEvent(event, task)
		}
	}
	for event := range eventToTask {
		err := db.Transaction(
			context.TODO(), func(ctx *cus.TxContext) error {
				return PublishEventToOutboxWithPayload(ctx, event, 3)
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	for true {
		msg, open := <-eventChan
		if open {
			if _, exist := eventToTask[msg.Event][msg.Task]; !exist {
				close(eventChan)
				t.Fatal(msg, errors.New("msg repeat"))
			}
			delete(eventToTask[msg.Event], msg.Task)
			if len(eventToTask[msg.Event]) == 0 {
				delete(eventToTask, msg.Event)
			}
			if len(eventToTask) == 0 {
				if len(eventChan) > 0 {
					close(eventChan)
					t.Fatal(eventChan, errors.New("msg repeat"))
				}
				t.Log("finish")
				close(eventChan)
				return
			}
		} else {
			t.Fatal(errors.New("chan close"))
		}
	}
}

func TestCustomerProcessingTimeout(t *testing.T) {
	var count atomic.Int32
	task := manager.Task(t.Name())
	msgChan := make(chan struct{})
	taskManage.Subscribe(
		task, func(payload []byte) error {
			count.Add(1)
			time.Sleep(time.Second * 20)
			msgChan <- struct{}{}
			return nil
		},
	)

	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			err := db.Transaction(
				context.TODO(), func(ctx *cus.TxContext) error {
					return PublishTaskToOutbox(ctx, task)
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			<-msgChan
			close(msgChan)
			if count.Load() != 1 {
				t.Fatal("msg repeat, repeat count:", count.Load()-1)
			}
		},
	)
}
