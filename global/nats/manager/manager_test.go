package manager

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSubscribeAndPublish(t *testing.T) {
	t.Parallel()
	var task = Task(t.Name())
	var count int
	manager := taskManage
	manager.Subscribe(
		task, func(payload []byte) error {
			if !reflect.DeepEqual(payload, []byte("1")) {
				t.Fail()
			}
			count++
			return nil
		},
	)
	time.Sleep(time.Second * 3)
	isSuccess := manager.Publish(task, []byte("1"))
	time.Sleep(time.Second * 3)
	if !isSuccess || count != 1 {
		t.Fail()
	}
}

func TestEventSubscribeAndPublish(t *testing.T) {
	t.Parallel()
	var event Event = Event(t.Name())
	var taskPrefix Task = Task(t.Name())

	taskM := taskManage
	eventM := eventManage
	var taskMap map[Task]bool
	taskMap = make(map[Task]bool)
	for i := 1; i <= 100; i++ {
		taskMap[taskPrefix+Task("_"+strconv.FormatInt(int64(i), 10))] = false
	}

	for task := range taskMap {
		// 订阅任务
		taskM.Subscribe(
			task, func(payload []byte) error {
				taskMap[task] = true
				return nil
			},
		)
		// 订阅事件触发任务
		eventM.Subscribe(event, task, func(eventData []byte) ([]byte, error) { return eventData, nil })
	}
	time.Sleep(time.Second * 1)
	// 发布事件
	eventM.Publish(event, []byte("test"))
	time.Sleep(time.Second * 20)
	for task, b := range taskMap {
		if !b {
			t.Fatal(task, "fail")
		}
	}
	t.Log("task trigger info", taskMap)
}

func TestDql(t *testing.T) {
	t.Parallel()
	taskM := taskManage
	var task Task = Task(t.Name())
	var count = 1
	taskM.Subscribe(
		task, func(payload []byte) error {
			count++
			return errors.New("test dql")
		},
	)
	time.Sleep(time.Second)
	taskM.Publish(task, []byte("test"))
	time.Sleep(time.Second * 30)
	batch, err := dlqManage.consumer.Fetch(10)
	if err != nil {
		t.Error(err)
	}
	for msg := range batch.Messages() {
		err = msg.Ack()
		if err != nil {
			t.Error(err)
		}
	}
}

func TestDqlRepublish(t *testing.T) {
	t.Parallel()
	taskM := taskManage
	var task Task = Task(t.Name())
	var count = 1
	taskM.Subscribe(
		task, func(payload []byte) error {
			count++
			return errors.New("test dql")
		},
	)
	time.Sleep(time.Second)
	for i := 0; i < 100; i++ {
		taskM.Publish(task, []byte("test_"+strconv.FormatInt(int64(i), 10)))
	}
	time.Sleep(time.Second * 10)
	t.Run(
		"republish die msg", func(t *testing.T) {
			taskM.Subscribe(
				task, func(payload []byte) error {
					count--
					return nil
				},
			)
			err := dlqManage.RepublishBatch(10, context.TODO())
			if err != nil {
				t.Error(err)
			}
		},
	)
	time.Sleep(time.Second * 5)
}

func BenchmarkDql(b *testing.B) {
	taskM := taskManage
	var task Task = Task(uuid.NewString())
	var count = b.N
	taskM.Subscribe(
		task, func(payload []byte) error {
			return errors.New("test dql")
		},
	)
	time.Sleep(time.Second * 5)
	for i := 0; i < b.N; i++ {
		taskM.Publish(task, []byte("test_"+strconv.FormatInt(int64(i), 10)))
	}
	time.Sleep(time.Second * 20)
	b.Run(
		"republish", func(b *testing.B) {
			taskM.Subscribe(
				task, func(payload []byte) error {
					count--
					return nil
				},
			)

			err := dlqManage.RepublishBatch(b.N, context.Background())
			if err != nil {
				b.Error(err)
			}
		},
	)
	time.Sleep(time.Second * 20)
	if count != 0 {
		b.Fatal("msg lose Publish:", b.N, " republish:", count)
	}
}
