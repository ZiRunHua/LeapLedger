package nats

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

func getTaskManage() *taskManager {
	var m taskManager
	err := m.init(js, taskLogger)
	if err != nil {
		panic(err)
	}
	return &m
}

func getEventManage(t *taskManager) *eventManager {
	var m eventManager
	err := m.init(js, t, eventLogger)
	if err != nil {
		panic(err)
	}
	return &m
}

func TestSubscribeAndPublish(t *testing.T) {
	var task Task = Task(t.Name())
	var count int
	manager := getTaskManage()
	manager.subscribe(
		task, func(msg jetstream.Msg) error {
			if !reflect.DeepEqual(msg.Data(), []byte("1")) {
				t.Fail()
			}
			count++
			return nil
		},
	)
	time.Sleep(time.Second * 3)
	isSuccess := manager.publish(task, []byte("1"))
	if !isSuccess || count != 1 {
		t.Fail()
	}
}

func TestLoadBalancing(t *testing.T) {
	var task Task = Task(t.Name())
	var g sync.WaitGroup
	var publishCount = 1000
	g.Add(publishCount)
	nodeSubscribe := func() {
		m := getTaskManage()
		m.subscribe(
			Task(t.Name()), func(msg jetstream.Msg) error {
				g.Done()
				return nil
			},
		)
		g.Wait()
	}
	go nodeSubscribe()
	go nodeSubscribe()
	go nodeSubscribe()
	go nodeSubscribe()
	time.Sleep(time.Second)
	m := getTaskManage()
	for i := 0; i < publishCount; i++ {
		m.publish(task, []byte{byte(i + 1)})
	}
	g.Wait()
}

func TestManager3(t *testing.T) {
	var task Task = Task(t.Name())
	var handleCount int

	m := getTaskManage()
	m.subscribe(
		task, func(msg jetstream.Msg) error {
			fmt.Printf("%s %s\n", msg.Subject(), msg.Data())
			handleCount++
			return errors.New("test")
		},
	)
	m.publish(task, []byte("test"))
	time.Sleep(time.Second * 10)
	t.Log(handleCount)
}

func TestEventSubscribeAndPublish(t *testing.T) {
	var event Event = Event(t.Name())
	var taskPrefix Task = Task(t.Name())

	taskM := getTaskManage()
	eventM := getEventManage(taskM)
	var tasks map[Task]bool
	tasks = make(map[Task]bool)
	for i := 1; i <= 100; i++ {
		tasks[taskPrefix+Task("_"+strconv.FormatInt(int64(i), 10))] = false
	}

	for ts, _ := range tasks {
		var task = ts
		// 订阅任务
		taskM.subscribe(
			task, func(msg jetstream.Msg) error {
				tasks[task] = true
				t.Log(task, "trigger")
				return nil
			},
		)
		// 订阅事件触发任务
		eventM.subscribe(
			event,
			task, func(eventData []byte) ([]byte, error) { return eventData, nil },
		)
	}
	// 发布事件
	eventM.publish(event, []byte("test"))
	time.Sleep(time.Second * 10)
	for task, b := range tasks {
		if !b {
			t.Fatal(task, "fail")
		}
	}
}

func TestDql(t *testing.T) {
	taskM := getTaskManage()
	var task Task = Task(t.Name())
	var count = 1
	taskM.subscribe(
		task, func(msg jetstream.Msg) error {
			t.Log(count, msg.Headers().Get(msgHeaderKeySubject), "handle msg fail")
			count++
			return errors.New("test dql")
		},
	)
	time.Sleep(time.Second)
	taskM.publish(task, []byte("test"))
	time.Sleep(time.Second * 30)
	msgs, err := dlqManage.consumer.Fetch(10)
	if err != nil {
		t.Error(err)
	}
	for msg := range msgs.Messages() {
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
	taskM.subscribe(
		task, func(msg jetstream.Msg) error {
			count++
			return errors.New("test dql")
		},
	)
	time.Sleep(time.Second)
	for i := 0; i < 100; i++ {
		taskM.publish(task, []byte("test_"+strconv.FormatInt(int64(i), 10)))
	}
	time.Sleep(time.Second * 10)
	t.Run("republish die msg", func(t *testing.T) {
		taskM.subscribe(
			task, func(msg jetstream.Msg) error {
				count--
				return nil
			},
		)
		err := dlqManage.republishBatch(10, context.TODO())
		if err != nil {
			t.Error(err)
		}
	})
	time.Sleep(time.Second * 5)
}

func BenchmarkDql(b *testing.B) {
	taskM := taskManage
	var task Task = Task(uuid.NewString())
	var count = b.N
	taskM.subscribe(
		task, func(msg jetstream.Msg) error {
			return errors.New("test dql")
		},
	)
	time.Sleep(time.Second * 5)
	for i := 0; i < b.N; i++ {
		taskM.publish(task, []byte("test_"+strconv.FormatInt(int64(i), 10)))
	}
	time.Sleep(time.Second * 20)
	b.Run("republish", func(b *testing.B) {
		taskM.subscribe(
			task, func(msg jetstream.Msg) error {
				count--
				return nil
			},
		)

		err := dlqManage.republishBatch(b.N, context.Background())
		if err != nil {
			b.Error(err)
		}
	})
	time.Sleep(time.Second * 20)
	if count != 0 {
		b.Fatal("msg lose publish :", b.N, " republish:", count)
	}
}
