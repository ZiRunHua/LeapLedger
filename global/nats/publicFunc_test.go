package nats

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
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
	success := false
	SubscribeTask(
		task, func(ctx context.Context) error {
			success = true
			return nil
		},
	)
	t.Run(
		"Publish task", func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Second * 3)
			if !PublishTask(task) {
				t.Error("Publish fail")
			}
			time.Sleep(time.Second * 3)
			if !success {
				t.Fail()
			}
		},
	)
	withPayloadTask, data := Task(t.Name()+uuid.NewString()), newTaskData()
	var withPayloadTaskSuccess bool
	SubscribeTaskWithPayloadAndProcessInTransaction(
		withPayloadTask, func(pushData taskData, ctx context.Context) error {
			withPayloadTaskSuccess = reflect.DeepEqual(data, pushData)
			return nil
		},
	)
	t.Run(
		"Publish task With payload", func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Second * 3)
			if !PublishTaskWithPayload(withPayloadTask, data) {
				t.Error("Publish fail")
			}
			time.Sleep(time.Second * 3)
			if !withPayloadTaskSuccess {
				t.Fail()
			}
		},
	)
}

func TestEventPublishAndSubscribe(t *testing.T) {
	taskMap, event := make(map[Task]int), Event(uuid.NewString())
	for i := 0; i < 10; i++ {
		task := Task("task_" + uuid.NewString())
		taskMap[task] = 0
		SubscribeTask(
			task, func(ctx context.Context) error {
				taskMap[task]++
				return nil
			},
		)
	}
	time.Sleep(time.Second)
	for task := range taskMap {
		BindTaskToEvent(event, task)
	}
	t.Run(
		"publish", func(t *testing.T) {
			time.Sleep(time.Second * 3)
			PublishEvent(event)
			time.Sleep(time.Second * 10)
			for task, value := range taskMap {
				if value != 1 {
					t.Log(task, "fail trigger count", value)
				}
			}
			t.Log("task trigger info", taskMap)
		},
	)
}

func TestSubscribeEvent(t *testing.T) {
	name := t.Name() + uuid.NewString()
	event := Event(name)
	var count atomic.Int32
	count.Add(10)
	var retryCount atomic.Int32
	SubscribeEvent(
		event, name, func(v int, ctx context.Context) error {
			if retryCount.Add(1) < 10 {
				return errors.New("test retry")
			}
			count.Add(int32(v))
			return nil
		},
	)
	task := Task(name + "_task_1")
	BindTaskToEvent(event, task)
	SubscribeTaskWithPayload(
		task, func(v int, ctx context.Context) error {
			if retryCount.Add(1) < 10 {
				return errors.New("test retry")
			}
			count.Add(int32(v))
			return nil
		},
	)
	task = Task(name + "_task_2")
	BindTaskToEvent(event, task)
	SubscribeTaskWithPayload(
		task, func(v int, ctx context.Context) error {
			if retryCount.Add(1) < 10 {
				return errors.New("test retry")
			}
			count.Add(int32(v))
			return nil
		},
	)
	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			time.Sleep(3 * time.Second)
			PublishEventWithPayload(event, 1)
			time.Sleep(30 * time.Second)
			if count.Load() != 13 {
				t.Fatal(count.Load())
			}
		},
	)
}

func TestOutboxTask(t *testing.T) {
	taskMap := make(map[Task]*atomic.Int32)
	var retryCount int32 = 2
	for i := 0; i < 3; i++ {
		task := Task(t.Name() + "task_" + uuid.NewString())
		taskMap[task] = new(atomic.Int32)
		SubscribeTaskWithPayload(
			task, func(data int32, ctx context.Context) error {
				if taskMap[task].Add(-1) > -retryCount {
					return errors.New("test retry")
				}
				taskMap[task].Add(data)
				return nil
			},
		)
	}
	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Second * 3)
			for task := range taskMap {
				err := db.Transaction(
					context.TODO(), func(ctx *cus.TxContext) error {
						return PublishTaskToOutboxWithPayload(ctx, task, retryCount+1)
					},
				)
				if err != nil {
					t.Fatal(err)
				}
			}
			time.Sleep(time.Second * 30)
			for task, i := range taskMap {
				if i.Load() != 1 {
					t.Fatal(task, i)
				}
			}
		},
	)
}

func TestOutboxEvent(t *testing.T) {
	eventMap := make(map[Event]*atomic.Int32)
	eventToTask := make(map[Event][]Task)
	for i := 0; i < 10; i++ {
		event := Event("event_" + uuid.NewString())
		eventMap[event] = &atomic.Int32{}
		for j := 0; j < 3; j++ {
			task := Task("task_" + uuid.NewString())
			eventToTask[event] = append(eventToTask[event], task)
			SubscribeTaskWithPayload(
				task, func(t int, ctx context.Context) error {
					eventMap[event].Add(3)
					return nil
				},
			)
			BindTaskToEvent(event, task)
		}
	}
	t.Run(
		"public", func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Second * 3)
			for event := range eventMap {
				err := db.Transaction(
					context.TODO(), func(ctx *cus.TxContext) error {
						return PublishEventToOutboxWithPayload(ctx, event, 3)
					},
				)
				if err != nil {
					t.Fatal(err)
				}
			}
			time.Sleep(time.Second * 20)
			for event, num := range eventMap {
				if num.Load() != 9 {
					t.Fatal(event, num.Load())
				}
			}
		},
	)
}

func TestCustomerProcessingTimeout(t *testing.T) {
	var count atomic.Int32
	task := manager.Task(t.Name())
	taskManage.Subscribe(
		task, func(payload []byte) error {
			count.Add(1)
			time.Sleep(time.Second * 20)
			return nil
		},
	)

	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			err := db.Transaction(
				context.TODO(), func(ctx *cus.TxContext) error {
					return PublishTaskToOutbox(ctx, Task(task))
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Second * 31)
			if count.Load() != 1 {
				t.Fatal("count not is 1,count:", count.Load())
			}
		},
	)
}
