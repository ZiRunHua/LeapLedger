package manager

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ZiRunHua/LeapLedger/util/rand"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func init() {
	UpdateTestBackOff()
}
func TestSubscribeAndPublish(t *testing.T) {
	var taskPrefix = Task(t.Name() + uuid.NewString())
	taskList := []struct {
		Task Task
		Msg  []byte
	}{
		{
			taskPrefix + "_1", []byte("msg1"),
		},
		{
			taskPrefix + "_2", []byte("123"),
		},
		{
			taskPrefix + "_3", []byte("msg3"),
		},
	}
	msgChan := make(chan []byte, len(taskList))
	for _, task := range taskList {
		taskManage.Subscribe(
			task.Task, func(payload []byte) error {
				if !reflect.DeepEqual(payload, task.Msg) {
					t.Fail()
				}
				msgChan <- payload
				return nil
			},
		)
	}
	t.Run(
		"publish", func(t *testing.T) {
			t.Parallel()
			for _, task := range taskList {
				taskManage.Publish(task.Task, task.Msg)
			}
			count := 0
			for true {
				count++
				if count == len(taskList) {
					if len(msgChan) == 0 {
						break
					} else {
						t.Fatal()
					}
				}
			}
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
	msgChan := make(chan Task)
	for task := range taskMap {
		taskManage.Subscribe(
			task, func(payload []byte) error {
				msgChan <- task
				return nil
			},
		)
		eventManage.Subscribe(event, task, func(eventData []byte) ([]byte, error) { return eventData, nil })
	}
	t.Run(
		"event publish", func(t *testing.T) {
			t.Parallel()
			eventManage.Publish(event, []byte("test"))
			count := 0
			for true {
				taskMap[<-msgChan] = true
				count++
				if count == len(taskMap) {
					if len(msgChan) == 0 {
						for task, result := range taskMap {
							if !result {
								t.Fatal(task, " not trigger")
							}
						}
						break
					} else {
						t.Fatal()
					}
				}
			}
		},
	)
}

func TestDql(t *testing.T) {
	var task, event = Task(t.Name() + rand.String(12)), Event(t.Name() + rand.String(12))
	msgChan, msg := make(chan []byte), []byte(rand.String(12))
	taskManage.Subscribe(
		task, func(payload []byte) error {
			msgChan <- payload
			return errors.New("test dql")
		},
	)
	taskManage.Publish(task, msg)
	count := 0
	for true {
		<-msgChan
		count++
		if count == len(backOff)+1 {
			if len(msgChan) == 0 {
				break
			} else {
				t.Fatal("try too many times :", count+len(msgChan))
			}
		}
	}
	taskManage.Subscribe(
		task, func(payload []byte) error {
			msgChan <- payload
			return nil
		},
	)
	time.Sleep(backOff[len(backOff)-1])
	_, err := dlqManage.RepublishBatch(1, context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	republishPayload := <-msgChan
	if !reflect.DeepEqual(republishPayload, msg) {
		t.Fatal("reConsume payload not compare", string(republishPayload), msg)
	}
	t.Run(
		"event and task dql", func(t *testing.T) {
			type TestEvent struct {
				Event        Event
				EventData    []byte
				TriggerTasks []struct {
					task          Task
					fetchTaskData func(eventData []byte) ([]byte, error)
					retryCount    int
				}
				ExecConsumers []struct {
					name       string
					retryCount int
				}
			}
			var testEvent = TestEvent{
				Event:     event + Event(rand.String(12)),
				EventData: []byte("event_data_" + rand.String(12)),
				TriggerTasks: []struct {
					task          Task
					fetchTaskData func(eventData []byte) ([]byte, error)
					retryCount    int
				}{
					{
						task:          task + "_0",
						fetchTaskData: func(eventData []byte) ([]byte, error) { return eventData, nil },
						retryCount:    0,
					},
					{
						task:          task + "_1",
						fetchTaskData: func(eventData []byte) ([]byte, error) { return eventData, nil },
						retryCount:    (len(backOff)) / 2,
					},
					{
						task:          task + "_2",
						fetchTaskData: func(eventData []byte) ([]byte, error) { return eventData, nil },
						retryCount:    len(backOff) + 1,
					},
				},
				ExecConsumers: []struct {
					name       string
					retryCount int
				}{
					{
						name:       string(task) + "_4",
						retryCount: len(backOff) / 2,
					},
					{
						name:       string(task) + "_5",
						retryCount: len(backOff) + 1,
					},
					{
						name:       string(task) + "_3",
						retryCount: 0,
					},
				},
			}

			msgChan = make(chan []byte, 10)
			var needToReConsume, noNeedToReConsume int
			for _, triggerTask := range testEvent.TriggerTasks {
				retryCount := 0
				taskManage.Subscribe(
					triggerTask.task, func(payload []byte) error {
						if retryCount == triggerTask.retryCount {
							t.Log(triggerTask.task, "finish")
							msgChan <- payload
							return nil
						} else if retryCount > triggerTask.retryCount {
							t.Fatal(
								"too many retries or re-consuming errors,task:", triggerTask.task, retryCount, ">",
								triggerTask.retryCount,
							)
							return nil
						}
						retryCount++
						return errors.New("test dql")
					},
				)
				eventManage.Subscribe(
					testEvent.Event, triggerTask.task, func(eventData []byte) ([]byte, error) { return eventData, nil },
				)
				if triggerTask.retryCount >= len(backOff)+1 {
					needToReConsume++
				} else {
					noNeedToReConsume++
				}
			}

			for _, execConsumer := range testEvent.ExecConsumers {
				retryCount := 0
				eventManage.SubscribeToNewConsumer(
					testEvent.Event,
					execConsumer.name, func(payload []byte) error {
						if retryCount == execConsumer.retryCount {
							t.Log(execConsumer.name, "finish")
							msgChan <- payload
							return nil
						} else if retryCount > execConsumer.retryCount {
							t.Fatal(
								"too many retries or re-consuming errors,consumer:", execConsumer.name, retryCount,
								">",
								execConsumer.retryCount,
							)
							return nil
						}
						retryCount++
						return errors.New("test dql")
					},
				)
				if execConsumer.retryCount >= len(backOff)+1 {
					needToReConsume++
				} else {
					noNeedToReConsume++
				}
			}
			eventManage.Publish(testEvent.Event, testEvent.EventData)
			var successMsg int
			for true {
				msg = <-msgChan
				successMsg++
				if !reflect.DeepEqual(msg, testEvent.EventData) {
					t.Fatal("payload not equal", msg, testEvent.EventData)
				}
				if successMsg >= noNeedToReConsume {
					if noNeedToReConsume == successMsg && len(msgChan) == 0 {
						break
					}
					t.Fatal("success msg to much:", noNeedToReConsume, successMsg+len(msgChan))
				}
			}
			t.Run(
				"test reConsume", func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
					defer cancel()
					g, _ := errgroup.WithContext(ctx)
					g.Go(
						func() error {
							time.Sleep(backOff[len(backOff)-1])
							for true {
								count, err := dlqManage.RepublishBatch(10, context.TODO())
								if err != nil {
									t.Fatal(err)
								}
								if count == 0 {
									break
								}
							}
							return nil
						},
					)
					g.Go(
						func() error {
							successMsg = 0
							for true {
								msg = <-msgChan
								successMsg++
								if !reflect.DeepEqual(msg, testEvent.EventData) {
									t.Fatal("payload not equal", msg, testEvent.EventData)
								}
								if successMsg >= needToReConsume {
									if needToReConsume == successMsg && len(msgChan) == 0 {
										return nil
									}
									t.Fatal("success msg to much:", needToReConsume, successMsg, len(msgChan))
								}
							}
							t.Fatal("success msg to much:", needToReConsume, successMsg, len(msgChan))
							return nil
						},
					)
					err := g.Wait()
					if err != nil {
						t.Fatal(err)
					}
				},
			)
		},
	)
}
