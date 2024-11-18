package manager

// dead letter queue
import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"

	natsServer "github.com/nats-io/nats-server/v2/server"

	"context"
	"github.com/ZiRunHua/LeapLedger/util/dataTool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// The DlqManager is used to manage dead letters.
// Use a dead letter queue by passing in a jetstream.Stream registration.
// Messages in DlqManager will use stream_seq to query the complete message on the registered stream,
// provided that the message still exists, and the message will be published as a new message on the registered stream.
type (
	DlqManager interface {
		RepublishBatch(batch int, ctx context.Context) (msgNum int, err error)
	}

	dlqManager struct {
		DlqManager

		manageInitializers
		register     dlqStreamRegister
		pullConsumer pullConsumer
	}

	dlqRegisterStream interface {
		getStreamName() string
		reConsume(ctx context.Context, consumer string, streamSeq uint64) error
	}
)

const (
	dlqName   = "dlq"
	dlqPrefix = "dlq"
)

var (
	dlqLogPath = filepath.Join(natsLogPath, "dlq.log")
)

func (dm *dlqManager) init(
	js jetstream.JetStream, registerStreams []dlqRegisterStream, logger *zap.Logger) (err error) {
	dm.register.streamMap = dataTool.NewSyncMap[string, dlqRegisterStream]()

	err = dm.register.register(registerStreams...)
	if err != nil {
		return err
	}
	subjects, err := dm.register.getMaxDeliveriesEvents()
	if err != nil {
		return err
	}
	streamConfig := jetstream.StreamConfig{
		Name:      dlqName,
		Subjects:  subjects,
		Retention: jetstream.InterestPolicy,
	}
	consumerConfig := jetstream.ConsumerConfig{
		Name:       dlqPrefix + "_consumer",
		Durable:    dlqPrefix + "_consumer",
		AckPolicy:  jetstream.AckExplicitPolicy,
		MaxDeliver: 0,
	}
	err = dm.manageInitializers.init(js, streamConfig, consumerConfig, logger)
	if err != nil {
		return err
	}
	err = dm.pullConsumer.updateConfig(
		js, streamConfig.Name,
		jetstream.ConsumerConfig{
			Name:       dlqPrefix + "_pull_consumer",
			Durable:    dlqPrefix + "_pull_consumer",
			AckPolicy:  jetstream.AckExplicitPolicy,
			MaxDeliver: 0,
		},
	)
	if err != nil {
		return err
	}
	return dm.manageInitializers.setMainConsumerConsume(
		context.TODO(), func(subject string, payload []byte) error {
			dm.logger.Info(
				"msg", zap.String("subject", subject),
				zap.ByteString("data", payload),
			)
			return nil
		},
	)
}

func (dm *dlqManager) RepublishBatch(batch int, ctx context.Context) (msgNum int, err error) {
	msgBatch, err := dm.pullConsumer.fetchMsg(batch)
	if err != nil {
		if errors.Is(err, nats.ErrMsgNotFound) {
			return 0, nil
		}
		return 0, err
	}

	for msg := range msgBatch.Messages() {
		msgNum++
		dm.republishDieMsg(ctx, msg)
	}
	return msgNum, err
}

func (dm *dlqManager) republishDieMsg(ctx context.Context, msg jetstream.Msg) (isSuccess bool) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			dm.logger.Panic(
				"reConsume", zap.String("subject", msg.Subject()), zap.ByteString("data", msg.Data()),
				zap.Any("panic", r), zap.Stack(string(debug.Stack())),
			)
			return
		}
		if err == nil {
			err = msg.Ack()
		}
		if err != nil {
			dm.logger.Error(
				"reConsume", zap.String("subject", msg.Subject()), zap.ByteString("data", msg.Data()),
				zap.Error(err),
			)
		}
	}()
	var advisory natsServer.JSConsumerDeliveryExceededAdvisory
	err = json.Unmarshal(msg.Data(), &advisory)
	if err != nil {
		return
	}
	stream, exist := dm.register.streamMap.Load(advisory.Stream)
	if !exist {
		err = ErrStreamNotExist
		return
	}
	err = stream.reConsume(ctx, advisory.Consumer, advisory.StreamSeq)
	if errors.Is(err, jetstream.ErrMsgNotFound) {
		// ignore
		err = nil
		return false
	}
	return err == nil
}

type dlqStreamRegister struct {
	streamMap dataTool.Map[string, dlqRegisterStream]
}

func (dsr *dlqStreamRegister) register(streams ...dlqRegisterStream) error {
	for _, stream := range streams {
		dsr.streamMap.LoadOrStore(stream.getStreamName(), stream)
	}
	return nil
}

func (dsr *dlqStreamRegister) getMaxDeliveriesEvents() ([]string, error) {
	var events []string
	dsr.streamMap.Range(
		func(_ string, stream dlqRegisterStream) bool {
			events = append(
				events, fmt.Sprintf(
					"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.%s.*",
					stream.getStreamName(),
				),
			)
			return true
		},
	)
	return events, nil
}
