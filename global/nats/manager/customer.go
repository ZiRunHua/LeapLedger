package manager

import (
	"context"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

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
	return mi.consumer.FetchNoWait(batch)
}
