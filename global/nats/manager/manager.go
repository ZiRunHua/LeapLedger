package manager

import (
	"context"
	"go.uber.org/zap"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const msgHeaderKeySubject = "subject"

type msgManger interface {
	receiveMsg(msg jetstream.Msg)
	getHandler(subject string) (MessageHandler, error)
	msgHandle(subject string, payload []byte) error
}

type MsgType interface {
	subject() string
	queue() string
}

type MessageHandler func(payload []byte) error

var backOff = []time.Duration{
	time.Second * 10,
	time.Second * 60,
	time.Second * 300,
	time.Hour,
	time.Hour * 7,
	time.Hour * 24,
}

var testBackOff = []time.Duration{
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
	time.Second,
}

var updateTestBackOffOnce sync.Once

// UpdateTestBackOff
// Most test samples are suspended for 30 seconds to wait for message consumption,
// and test whether retry and dead letter queues work properly,
// so to ensure that the test samples execute properly,
// you need to retry at least 10 times within 30 seconds.
func UpdateTestBackOff() {
	updateFunc := func() {
		ctx := context.TODO()
		backOff = testBackOff
		err := eventManage.updateAllConsumerConfig(
			func(config *jetstream.ConsumerConfig) error {
				config.BackOff = backOff
				config.MaxDeliver = len(backOff) + 1
				return nil
			}, ctx,
		)
		if err != nil {
			panic(err)
		}
		info, err := taskManage.consumer.Info(ctx)
		if err != nil {
			panic(err)
		}
		info.Config.BackOff = backOff
		info.Config.MaxDeliver = len(backOff) + 1
		_, err = taskManage.stream.UpdateConsumer(ctx, info.Config)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 3)
		log.Print("update test back off finish")
	}
	updateTestBackOffOnce.Do(updateFunc)
}

type manageInitializers struct {
	js             jetstream.JetStream
	stream         jetstream.Stream
	consumer       jetstream.Consumer
	consumerManger ConsumerManger

	logger *zap.Logger
}

func (mi *manageInitializers) init(
	js jetstream.JetStream, streamConfig jetstream.StreamConfig, consumerConfig jetstream.ConsumerConfig,
	logger *zap.Logger,
) (err error) {
	mi.js, mi.logger = js, logger
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mi.stream, err = mi.js.CreateOrUpdateStream(ctx, streamConfig)
	if err != nil {
		return err
	}
	mi.consumer, err = mi.stream.CreateOrUpdateConsumer(ctx, consumerConfig)
	if err != nil {
		return err
	}
	mi.consumerManger, err = NewConsumerManger(ctx, mi.stream, mi.consumer, logger)
	if err != nil {
		return err
	}
	return nil
}

func (mi *manageInitializers) setMainConsumerConsume(ctx context.Context, handler consumerMessageHandler) (err error) {
	return mi.consumerManger.Consume(ctx, mi.consumer, handler)
}
