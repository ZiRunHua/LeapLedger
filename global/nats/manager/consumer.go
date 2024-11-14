package manager

import (
	"context"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

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
