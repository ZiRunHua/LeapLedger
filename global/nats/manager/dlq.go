package manager

// dead letter queue
import (
	"errors"
	"fmt"

	"github.com/ZiRunHua/LeapLedger/util/dataTool"
	natsServer "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

import (
	"context"
	"encoding/json"
)

// The DlqManager is used to manage dead letters.
// Use a dead letter queue by passing in a jetstream.Stream registration.
// Messages in DlqManager will use stream_seq to query the complete message on the registered stream,
// provided that the message still exists, and the message will be published as a new message on the registered stream.
type DlqManager interface {
	RepublishBatch(batch int, ctx context.Context) (int, error)
}

const dlqName = "dlq"
const dlqPrefix = "dlq"
const dlqLogPath = natsLogPath + "dlq.log"

type dlqManager struct {
	DlqManager

	manageInitializers
	dlqMsgHandler
	register     dlqStreamRegister
	pullCustomer pullCustomer
}

func (dm *dlqManager) init(js jetstream.JetStream, registerStream []jetstream.Stream, logger *zap.Logger) (err error) {
	dm.logger, dm.register.streamMap = logger, dataTool.NewSyncMap[string, jetstream.Stream]()

	err = dm.register.register(registerStream...)
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
	customerConfig := jetstream.ConsumerConfig{
		Name:       dlqPrefix + "_customer",
		Durable:    dlqPrefix + "_customer",
		AckPolicy:  jetstream.AckExplicitPolicy,
		MaxDeliver: 0,
	}
	err = dm.manageInitializers.init(js, streamConfig, customerConfig)
	if err != nil {
		return err
	}
	err = dm.pullCustomer.updateConfig(
		js, streamConfig.Name,
		jetstream.ConsumerConfig{
			Name:       dlqPrefix + "_pull_customer",
			Durable:    dlqPrefix + "_pull_customer",
			AckPolicy:  jetstream.AckExplicitPolicy,
			MaxDeliver: 0,
		},
	)
	if err != nil {
		return err
	}
	_, err = dm.consumer.Consume(dm.receiveMsg)
	return err
}

func (dm *dlqManager) RepublishBatch(batch int, ctx context.Context) (int, error) {
	msgBatch, err := dm.pullCustomer.fetchMsg(batch)
	if err != nil {
		if errors.Is(err, nats.ErrMsgNotFound) {
			return 0, nil
		}
		return 0, err
	}
	var count int
	for msg := range msgBatch.Messages() {
		count++
		err = dm.republishDieMsg(msg, ctx)
		if err != nil {
			dm.logger.Error("republishDieMsg err", zap.Error(err))
			err = msg.Nak()
			if err != nil {
				dm.logger.Error("Republish nck", zap.Error(err))
			}
		} else {
			err = msg.Ack()
			if err != nil {
				dm.logger.Error("Republish ack", zap.Error(err))
			}
		}
	}
	return count, nil
}

func (dm *dlqManager) republishDieMsg(msg jetstream.Msg, ctx context.Context) (err error) {
	var advisory natsServer.JSConsumerDeliveryExceededAdvisory
	err = json.Unmarshal(msg.Data(), &advisory)
	if err != nil {
		return err
	}
	var republishMsg *nats.Msg
	republishMsg, err = dm.getMsgByAdvisory(advisory, ctx)
	if err != nil {
		return err
	}
	_, err = dm.js.PublishMsg(ctx, republishMsg)
	return err
}

func (dm *dlqManager) getMsgByAdvisory(advisory natsServer.JSConsumerDeliveryExceededAdvisory, ctx context.Context) (
	*nats.Msg, error,
) {
	streamRawMsg, err := dm.register.selectMsgByDeliveryExceededAdvisory(advisory, ctx)
	if err != nil {
		return nil, err
	}
	return &nats.Msg{
		Subject: streamRawMsg.Subject,
		Data:    streamRawMsg.Data,
		Header:  streamRawMsg.Header,
	}, nil
}

type dlqMsgHandler struct {
	logger *zap.Logger
}

func (dmh *dlqMsgHandler) receiveMsg(msg jetstream.Msg) {
	dmh.logMsg(msg)
	err := msg.Ack()
	if err != nil {
		dmh.logger.Error("receive msg", zap.Error(err))
	}
}

func (dmh *dlqMsgHandler) logMsg(msg jetstream.Msg) {
	dmh.logger.Info(
		"msg", zap.String(msgHeaderKeySubject, msg.Headers().Get(msgHeaderKeySubject)),
		zap.String("data", string(msg.Data())),
	)
}

type dlqStreamRegister struct {
	streamMap dataTool.Map[string, jetstream.Stream]
}

func (dsr *dlqStreamRegister) register(streams ...jetstream.Stream) error {
	for _, stream := range streams {
		dsr.streamMap.LoadOrStore(stream.CachedInfo().Config.Name, stream)
	}
	return nil
}

func (dsr *dlqStreamRegister) getMaxDeliveriesEvents() ([]string, error) {
	var events []string
	dsr.streamMap.Range(
		func(_ string, stream jetstream.Stream) bool {
			events = append(
				events, fmt.Sprintf(
					"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.%s.*",
					stream.(jetstream.Stream).CachedInfo().Config.Name,
				),
			)
			return true
		},
	)
	return events, nil
}

func (dsr *dlqStreamRegister) selectMsgByDeliveryExceededAdvisory(
	advisory natsServer.JSConsumerDeliveryExceededAdvisory,
	ctx context.Context,
) (*jetstream.RawStreamMsg, error) {
	stream, exist := dsr.streamMap.Load(advisory.Stream)
	if !exist {
		return nil, ErrStreamNotExist
	}
	return stream.(jetstream.Stream).GetMsg(ctx, advisory.StreamSeq)
}
