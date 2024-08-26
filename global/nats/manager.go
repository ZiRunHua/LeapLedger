package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const msgHeaderKeySubject = "subject"

type msgManger interface {
	receiveMsg(msg jetstream.Msg)
	getHandler(subject string) (MessageHandler, error)
	msgHandle(msg jetstream.Msg) error
}

type MsgType interface {
	subject() string
	queue() string
}
type MessageHandler func(msg jetstream.Msg) error

type Publisher interface {
	publish(task MsgType, payload []byte) bool
}

var backOff = []time.Duration{
	time.Millisecond * 250,
	time.Millisecond * 500,
	time.Second * 3,
	time.Second * 30,
	time.Second * 300,
	time.Hour,
	time.Hour * 7,
	time.Hour * 24,
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

type pullCustomer struct {
	consumer jetstream.Consumer
}

func (mi *pullCustomer) updateConfig(
	js jetstream.JetStream,
	streamName string,
	config jetstream.ConsumerConfig,
) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	mi.consumer, err = js.CreateOrUpdateConsumer(ctx, streamName, config)
	return err
}

func (mi *pullCustomer) fetchMsg(batch int) (jetstream.MessageBatch, error) {
	return mi.consumer.Fetch(batch)
}
