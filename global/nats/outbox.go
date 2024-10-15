package nats

import (
	"KeepAccount/global/db"
	"KeepAccount/global/nats/manager"
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type outbox struct {
	Id       uint
	ExecType outboxType
	Type     string
	Payload  []byte
	failErr  string
	gorm.Model
}
type outboxType string

const outboxTypeTask outboxType = "task"
const outboxTypeEvent outboxType = "event"

func (o *outbox) TableName() string {
	return "core_outbox"
}
func (o *outbox) fail(db *gorm.DB, err error) error {
	return db.Model(o).Update("fail_err", err.Error()).Error
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
	err := db.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).First(&o, id).Error
	return o, err
}

func (oServer *outboxServer) getHandleTransaction(t outboxType) handler[uint] {
	var getMsgHandler = oServer.getMessageHandler(t)
	return func(id uint, ctx context.Context) error {
		tx := db.Get(ctx)
		o, err := oServer.getOutboxAndLockById(tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		msgHandler, err := getMsgHandler(o.Type)
		if err != nil {

			return err
		}
		err = msgHandler(o.Payload)
		if err != nil {
			_ = o.fail(tx, err)
			return err
		}
		return o.completeReceipt(tx)
	}
}

func (oServer *outboxServer) getMessageHandler(t outboxType) func(string) (manager.MessageHandler, error) {
	switch t {
	case outboxTypeTask:
		return func(t string) (manager.MessageHandler, error) {
			return taskManage.GetMessageHandler(manager.Task(t))
		}
	case outboxTypeEvent:
		return func(t string) (manager.MessageHandler, error) {
			return eventManage.GetMessageHandler(manager.Event(t))
		}
	default:
		panic("error outboxType")
	}
}
