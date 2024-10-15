package nats

import (
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
	"KeepAccount/util/rand"
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

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
	t.Parallel()
	t.Run(
		"Publish task", func(t *testing.T) {
			task := Task(uuid.NewString())
			success := false
			SubscribeTask(
				task, func(ctx context.Context) error {
					success = true
					return nil
				},
			)
			time.Sleep(time.Second)
			if !PublishTask(task) {
				t.Error("Publish fail")
			}
			time.Sleep(time.Second)
			if !success {
				t.Fail()
			}
		},
	)
	t.Run(
		"Publish task With payload", func(t *testing.T) {
			task, data := Task(uuid.NewString()), newTaskData()
			var success bool
			SubscribeTaskWithPayloadAndProcessInTransaction[taskData](
				task, func(pushData taskData, ctx context.Context) error {
					success = reflect.DeepEqual(data, pushData)
					return nil
				},
			)
			time.Sleep(time.Second)
			if !PublishTaskWithPayload[taskData](task, data) {
				t.Error("Publish fail")
			}
			time.Sleep(time.Second * 3)
			if !success {
				t.Fail()
			}
		},
	)
}

func TestEventPublishAndSubscribe(t *testing.T) {
	t.Parallel()
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
	PublishEvent(event)
	time.Sleep(time.Second * 20)
	for task, value := range taskMap {
		if value != 1 {
			t.Log(task, "fail trigger count", value)
		}
	}
	t.Log("task trigger info", taskMap)
}

func TestSubscribeEvent(t *testing.T) {
	t.Parallel()
	event := Event(uuid.NewString())
	var count atomic.Int32
	count.Add(10)
	var retryCount atomic.Int32
	SubscribeEvent(
		event, string(event), func(t int, ctx context.Context) error {
			retryCount.Add(1)
			if retryCount.Load() < 10 {
				return errors.New("test retry")
			}
			count.Add(int32(t))
			return nil
		},
	)
	task := Task(uuid.NewString())
	BindTaskToEvent(event, task)
	SubscribeTaskWithPayload(
		task, func(t int, ctx context.Context) error {
			retryCount.Add(1)
			if retryCount.Load() < 10 {
				return errors.New("test retry")
			}
			count.Add(int32(t))
			return nil
		},
	)
	task = Task(uuid.NewString())
	BindTaskToEvent(event, task)
	SubscribeTaskWithPayload(
		task, func(t int, ctx context.Context) error {
			retryCount.Add(1)
			if retryCount.Load() < 10 {
				return errors.New("test retry")
			}
			count.Add(int32(t))
			return nil
		},
	)
	PublishEventWithPayload(event, 1)
	time.Sleep(20 * time.Second)
	if count.Load() != 13 {
		t.Fatal(count.Load())
	}
}

func TestOutboxTask(t *testing.T) {
	t.Parallel()
	taskMap := make(map[Task]int)
	for i := 0; i < 10; i++ {
		task := Task("task_" + uuid.NewString())
		taskMap[task] = 0
		SubscribeTaskWithPayload(
			task, func(data int, ctx context.Context) error {
				if rand.Int(10) > 6 {
					return errors.New("test retry")
				}
				taskMap[task] += data
				return nil
			},
		)
	}
	time.Sleep(time.Second)
	for task := range taskMap {
		err := db.Transaction(
			context.TODO(), func(ctx *cus.TxContext) error {
				return PublishTaskToOutboxWithPayload(ctx, task, 1)
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(time.Second * 20)
	for task, i := range taskMap {
		if i != 1 {
			t.Fatal(task, i)
		}
	}
}

func TestOutboxEvent(t *testing.T) {
	t.Parallel()
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
	time.Sleep(time.Second)
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
}

func TestCustomerProcessingTimeout(t *testing.T) {
	t.Parallel()
	var count atomic.Int32
	task := manager.Task(t.Name())
	taskManage.Subscribe(
		task, func(payload []byte) error {
			count.Add(1)
			time.Sleep(time.Second * 20)
			return nil
		},
	)

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
}
