package nats

import "context"

func RepublishDieMsg(batch int, ctx context.Context) error {
	return dlqManage.RepublishBatch(batch, ctx)
}
