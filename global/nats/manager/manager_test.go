package manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

func TestSubscribeAndPublish(t *testing.T) {
	var task = Task(t.Name())
	var count int
	manager := taskManage
	manager.Subscribe(
		task, func(msg jetstream.Msg) error {
			if !reflect.DeepEqual(msg.Data(), []byte("1")) {
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

func TestManager(t *testing.T) {
	var task Task = Task(t.Name())
	var handleCount int

	m := taskManage
	m.Subscribe(
		task, func(msg jetstream.Msg) error {
			fmt.Printf("%s %s\n", msg.Subject(), msg.Data())
			handleCount++
			return errors.New("test")
		},
	)
	m.Publish(task, []byte("test"))
	time.Sleep(time.Second * 10)
	t.Log(handleCount)
}

func TestEventSubscribeAndPublish(t *testing.T) {
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
			task, func(msg jetstream.Msg) error {
				taskMap[task] = true
				return nil
			},
		)
		// 订阅事件触发任务
		eventM.Subscribe(event, task, func(eventData []byte) ([]byte, error) { return eventData, nil })
	}
	// 发布事件
	eventM.Publish(event, []byte("test"))
	time.Sleep(time.Second * 10)
	for task, b := range taskMap {
		if !b {
			t.Fatal(task, "fail")
		}
	}
	t.Log("task trigger info", taskMap)
}

func TestDql(t *testing.T) {
	taskM := taskManage
	var task Task = Task(t.Name())
	var count = 1
	taskM.Subscribe(
		task, func(msg jetstream.Msg) error {
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
	taskM := taskManage
	var task Task = Task(t.Name())
	var count = 1
	taskM.Subscribe(
		task, func(msg jetstream.Msg) error {
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
				task, func(msg jetstream.Msg) error {
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
		task, func(msg jetstream.Msg) error {
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
				task, func(msg jetstream.Msg) error {
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