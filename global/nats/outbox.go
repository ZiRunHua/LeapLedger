package nats

import (
	"context"
	"errors"
	"runtime/debug"

	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type outbox struct {
	Id       uint
	ExecType outboxType
	Type     string
	Payload  []byte `gorm:"type:TEXT"`
	FailErr  string `gorm:"type:TEXT"`
	gorm.Model
}
type outboxType string

const outboxTypeTask outboxType = "task"
const outboxTypeEvent outboxType = "event"

func (o *outbox) TableName() string {
	return "core_outbox"
}
func (o *outbox) fail(db *gorm.DB, errStr string) error {
	return db.Model(o).Update("fail_err", errStr).Error
}
func (o *outbox) completeReceipt(db *gorm.DB) error {
	return db.Delete(o).Error
}

var outboxService outboxServer

type outboxServer struct{}

func (oServer *outboxServer) sendToOutbox(db *gorm.DB, execType outboxType, t string, data []byte) (uint, error) {
	o := &outbox{ExecType: execType, Type: t, Payload: data}
	err := db.Create(o).Error
	if err != nil {
		return 0, err
	}
	return o.Id, nil
}

func (oServer *outboxServer) getOutboxAndLockById(db *gorm.DB, id uint) (outbox, error) {
	var o outbox
	err := db.Clauses(clause.Locking{Strength: "UPDATE", Options: clause.LockingOptionsNoWait}).First(&o, id).Error
	return o, err
}

func (oServer *outboxServer) getHandleTransaction(t outboxType) handler[uint] {
	var msgHandler func(execType string, payload []byte) error
	switch t {
	case outboxTypeTask:
		msgHandler = func(execType string, payload []byte) error {
			h, err := taskManage.GetMessageHandler(manager.Task(execType))
			if err != nil {
				return err
			}
			return h(payload)
		}
	case outboxTypeEvent:
		msgHandler = func(execType string, payload []byte) error {
			if eventManage.Publish(manager.Event(execType), payload) {
				return nil
			}
			return errors.New("fail publish event")
		}
	}
	return func(id uint, ctx context.Context) error {
		var msgHandleErr error
		err := db.Transaction(
			ctx, func(ctx *cus.TxContext) error {
				tx := ctx.GetDb()
				o, err := oServer.getOutboxAndLockById(tx, id)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil
					}
					return err
				}
				defer func() {
					r := recover()
					if r != nil {
						_ = o.fail(tx, string(debug.Stack()))
					}
				}()
				msgHandleErr = msgHandler(o.Type, o.Payload)
				if msgHandleErr != nil {
					return o.fail(tx, msgHandleErr.Error())
				}
				return o.completeReceipt(tx)
			},
		)
		if err != nil {
			return err
		}
		return msgHandleErr
	}
}
