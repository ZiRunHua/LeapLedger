package nats

import "context"

func RepublishDieMsg(batch int, ctx context.Context) error {
	_, err := dlqManage.RepublishBatch(batch, ctx)
	return err
}
