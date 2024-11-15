package manager

import (
	"context"
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

// UpdateTestBackOff
// Most test samples are suspended for 30 seconds to wait for message consumption,
// and test whether retry and dead letter queues work properly,
// so to ensure that the test samples execute properly,
// you need to retry at least 10 times within 30 seconds.

var once sync.Once

func UpdateTestBackOff() {
	updateFunc := func() {
		ctx := context.TODO()
		backOff = testBackOff
		err := eventManage.updateAllCustomerConfig(
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
	once.Do(updateFunc)
}

type manageInitializers struct {
	js       jetstream.JetStream
	stream   jetstream.Stream
	consumer jetstream.Consumer
}

func (mi *manageInitializers) init(
	js jetstream.JetStream, streamConfig jetstream.StreamConfig, customerConfig jetstream.ConsumerConfig,
) (err error) {
	mi.js = js
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = mi.updateStreamConfig(ctx, streamConfig)
	if err != nil {
		return err
	}
	return mi.updateCustomerConfig(ctx, customerConfig)
}

func (mi *manageInitializers) updateStreamConfig(
	ctx context.Context,
	streamConfig jetstream.StreamConfig,
) (err error) {
	mi.stream, err = mi.js.CreateOrUpdateStream(ctx, streamConfig)
	return err
}

func (mi *manageInitializers) updateCustomerConfig(
	ctx context.Context,
	customerConfig jetstream.ConsumerConfig,
) (err error) {
	mi.consumer, err = mi.js.CreateOrUpdateConsumer(ctx, mi.stream.CachedInfo().Config.Name, customerConfig)
	return err
}

func (mi *manageInitializers) getCustomerConfig(ctx context.Context) (config jetstream.ConsumerConfig, err error) {
	info, err := mi.consumer.Info(ctx)
	if err != nil {
		return config, err
	}
	config = info.Config
	return config, err
}

func (mi *manageInitializers) newCustomer(ctx context.Context, config jetstream.ConsumerConfig) (
	jetstream.Consumer, error,
) {
	return mi.js.CreateOrUpdateConsumer(ctx, mi.stream.CachedInfo().Config.Name, config)
}
