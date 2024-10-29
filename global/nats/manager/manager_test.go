package manager

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func init() {
	UpdateTestBackOff()
}
func TestSubscribeAndPublish(t *testing.T) {
	var task = Task(t.Name() + uuid.NewString())
	taskList := []Task{task + "_1", task + "_2", task + "_3"}
	var count atomic.Int32
	for _, name := range taskList {
		taskManage.Subscribe(
			name, func(payload []byte) error {
				if !reflect.DeepEqual(payload, []byte("1")) {
					t.Fail()
				}
				count.Add(1)
				return nil
			},
		)
	}
	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Second * 3)
			for _, name := range taskList {
				taskManage.Publish(name, []byte("1"))
			}
			time.Sleep(time.Second * 10)
			if count.Load() != int32(len(taskList)) {
				t.Fatal("count not is ", count.Load())
			}
			t.Log("count is:", count.Load())
		},
	)
}

func TestEventSubscribeAndPublish(t *testing.T) {
	var event = Event(t.Name() + uuid.NewString())
	var taskPrefix = Task(t.Name() + uuid.NewString())

	taskMap := make(map[Task]bool)
	for i := 1; i <= 3; i++ {
		taskMap[taskPrefix+Task("_"+strconv.Itoa(i))] = false
	}

	for task := range taskMap {
		taskManage.Subscribe(
			task, func(payload []byte) error {
				taskMap[task] = true
				return nil
			},
		)
		eventManage.Subscribe(event, task, func(eventData []byte) ([]byte, error) { return eventData, nil })
	}
	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Second * 3)
			eventManage.Publish(event, []byte("test"))
			time.Sleep(time.Second * 30)
			for task, b := range taskMap {
				if !b {
					t.Fatal(task, "fail")
				}
			}
			t.Log("task trigger info", taskMap)
		},
	)
}

func TestDql(t *testing.T) {
	taskM := taskManage
	var task = Task(t.Name())
	var retryCount = 1
	taskM.Subscribe(
		task, func(payload []byte) error {
			retryCount++
			return errors.New("test dql")
		},
	)
	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			time.Sleep(time.Second * 3)
			taskM.Publish(task, []byte("test"))
			var backOffTime time.Duration
			for _, duration := range backOff {
				backOffTime += duration
			}
			time.Sleep(time.Second*3 + backOffTime)
			batch, err := dlqManage.consumer.Fetch(10)
			if err != nil {
				t.Error(err)
			}
			var processCount int
			for msg := range batch.Messages() {
				processCount++
				err = msg.Ack()
				if err != nil {
					t.Error(err)
				}
			}
			t.Log("retry count", retryCount, "process count", processCount)
		},
	)
}

func TestDqlRepublish(t *testing.T) {
	var task = Task(t.Name() + uuid.NewString())
	var num = 3
	var retryCount atomic.Uint32
	taskManage.Subscribe(
		task, func(payload []byte) error {
			retryCount.Add(1)
			return errors.New("test dql")
		},
	)
	t.Run(
		"republish die msg", func(t *testing.T) {
			time.Sleep(time.Second)
			for i := 0; i < num; i++ {
				taskManage.Publish(task, []byte("test_"+strconv.FormatInt(int64(i), 10)))
			}
			var backOffTime time.Duration
			for _, duration := range backOff {
				backOffTime += duration
			}
			t.Log("sleep", backOffTime)
			time.Sleep(time.Second*3 + backOffTime)
			var count atomic.Int32
			count.Add(int32(num))
			taskManage.Subscribe(
				task, func(payload []byte) error {
					count.Add(-1)
					return nil
				},
			)
			time.Sleep(time.Second * 3)
			t.Log("retry count", retryCount.Load())
			_, err := dlqManage.RepublishBatch(num*10, context.TODO())

			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Second * 30)
			if 0 != count.Load() {
				t.Fatal("die msg Remaining:", count.Load())
			}
		},
	)
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

			_, err := dlqManage.RepublishBatch(b.N, context.Background())
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
