package nats

import (
	"context"
	"github.com/google/uuid"
	"reflect"
	"testing"
	"time"
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
