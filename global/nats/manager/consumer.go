package manager

import (
	"context"
	"errors"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ZiRunHua/LeapLedger/util/dataTool"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type (
	// ConsumerManger is used to manage the consumer group in the stream
	// It records the consumption method of the consumption group,
	// which will help with the retry of messages in the dead letter queue
	ConsumerManger interface {
		NewConsumer(context.Context, func(*jetstream.ConsumerConfig) error, consumerMessageHandler) (
			jetstream.Consumer, error,
		)
		Consume(context.Context, jetstream.Consumer, consumerMessageHandler) error
		ReConsume(ctx context.Context, consumerName string, msg *jetstream.RawStreamMsg) error
		UpdateAllConsumerConfig(func(*jetstream.ConsumerConfig) error, context.Context) error
	}

	consumerManger struct {
		stream jetstream.Stream
		// consumer is main consumer,other consumers will be based on the configuration of this consumer
		consumer               jetstream.Consumer
		consumerMessageHandler dataTool.Map[string, consumerMessageHandler]

		logger *zap.Logger
	}
	consumerMessageHandler func(subject string, payload []byte) error
)

func NewConsumerManger(_ context.Context, stream jetstream.Stream, consumer jetstream.Consumer, logger *zap.Logger) (
	ConsumerManger, error,
) {
	var cm consumerManger
	cm.stream, cm.consumer, cm.logger = stream, consumer, logger
	cm.consumerMessageHandler = dataTool.NewSyncMap[string, consumerMessageHandler]()
	return &cm, nil
}

func (cm *consumerManger) ReConsume(_ context.Context, consumerName string, msg *jetstream.RawStreamMsg) error {
	handler, exist := cm.consumerMessageHandler.Load(consumerName)
	if !exist {
		return errors.New("consumer message handler not found")
	}
	return handler(msg.Subject, msg.Data)
}

func (cm *consumerManger) createOrUpdateConsumer(ctx context.Context, config jetstream.ConsumerConfig) (
	consumer jetstream.Consumer, err error,
) {
	return cm.stream.CreateOrUpdateConsumer(ctx, config)
}

func (cm *consumerManger) Consume(
	ctx context.Context, consumer jetstream.Consumer, handle consumerMessageHandler,
) (err error) {
	info, err := consumer.Info(ctx)
	if err != nil {
		return err
	}
	_, err = consumer.Consume(cm.ReceiveMsg(info.Name, handle))
	if err != nil {
		return err
	}
	cm.consumerMessageHandler.Store(info.Name, handle)
	return nil
}

func (cm *consumerManger) ReceiveMsg(name string, handle consumerMessageHandler) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		var err error
		defer func() {
			if r := recover(); r != nil {
				cm.logger.Panic(
					msg.Subject(), zap.String("consumer", name), zap.ByteString("data", msg.Data()),
					zap.Any("panic", r), zap.Stack(string(debug.Stack())),
				)
				err = msg.Nak()
			} else if err != nil {
				cm.logger.Error(msg.Subject(), zap.String("consumer", name), zap.ByteString("data", msg.Data()),
					zap.Error(err))
				err = msg.Nak()
			} else {
				err = msg.Ack()
			}
			if err != nil {
				cm.logger.Error(msg.Subject(), zap.String("consumer", name), zap.ByteString("data", msg.Data()),
					zap.Error(err))
			}
		}()
		err = handle(msg.Subject(), msg.Data())
	}
}

func (cm *consumerManger) NewConsumer(
	ctx context.Context,
	setConfig func(*jetstream.ConsumerConfig) error,
	handler consumerMessageHandler) (jetstream.Consumer, error) {
	info, err := cm.consumer.Info(ctx)
	if err != nil {
		return nil, err
	}
	config := info.Config
	err = setConfig(&config)
	if err != nil {
		return nil, err
	}
	if strings.Compare(config.Name, info.Config.Name) == 0 ||
		strings.Compare(config.Durable, info.Config.Durable) == 0 {
		return nil, errors.New("new consumer has the same name as the main consumer")
	}
	consumer, err := cm.createOrUpdateConsumer(ctx, config)
	if err != nil {
		return nil, err
	}
	if handler == nil {
		handler = func(_ string, _ []byte) error { return nil }
	}
	_, err = consumer.Consume(cm.ReceiveMsg(config.Name, handler))
	if err != nil {
		return nil, err
	}
	cm.consumerMessageHandler.Store(config.Name, handler)
	return consumer, nil
}

func (cm *consumerManger) iterateConsumers(
	ctx context.Context,
) (func(yield func(*jetstream.ConsumerInfo) bool), error) {
	consumersList := cm.stream.ListConsumers(ctx)
	if err := consumersList.Err(); err != nil {
		return nil, err
	}
	return func(yield func(*jetstream.ConsumerInfo) bool) {
		for info := range consumersList.Info() {
			if !yield(info) {
				return
			}
		}
	}, nil
}

func (cm *consumerManger) UpdateAllConsumerConfig(
	handler func(*jetstream.ConsumerConfig) error, ctx context.Context,
) error {
	consumers, err := cm.iterateConsumers(ctx)
	if err != nil {
		return err
	}
	for info := range consumers {
		err = handler(&info.Config)
		if err != nil {
			return err
		}
		_, err = cm.createOrUpdateConsumer(ctx, info.Config)
		if err != nil {
			return err
		}
	}
	return nil
}

type pullConsumer struct {
	consumer jetstream.Consumer
}

func (mi *pullConsumer) updateConfig(
	js jetstream.JetStream,
	streamName string,
	config jetstream.ConsumerConfig,
) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	mi.consumer, err = js.CreateOrUpdateConsumer(ctx, streamName, config)
	return err
}

func (mi *pullConsumer) fetchMsg(batch int) (jetstream.MessageBatch, error) {
	return mi.consumer.FetchNoWait(batch)
}
