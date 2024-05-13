package nats

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/initialize"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var natsConn = initialize.Nats
var config = initialize.Config.Nats
var natsLogger = initialize.NatsLogger

const TaskStatisticUpdate constant.Subject = "statisticUpdate"
const TaskTransactionSync constant.Subject = "transactionSync"

func TransSubscribe[Data any](subj constant.Subject, handleFunc func(*gorm.DB, Data) error) {
	Subscribe[Data](
		subj, func(data Data) error {
			return global.GvaDb.Transaction(
				func(tx *gorm.DB) error {
					return handleFunc(tx, data)
				},
			)
		},
	)
}

func Subscribe[T any](subj constant.Subject, handleFunc func(T) error) {
	if natsConn == nil {
		return
	}

	if !config.CanSubscribe(subj) {
		return
	}

	_, err := natsConn.Subscribe(
		string(subj), func(msg *nats.Msg) {
			var t T
			if err := json.Unmarshal(msg.Data, &t); err != nil {
				natsLogger.Error(msg.Subject, zap.Error(err))
				return
			}
			err := handleFunc(t)
			if err != nil {
				natsLogger.Error(msg.Subject, zap.Error(err))
			}
		},
	)
	if err != nil {
		natsLogger.Error(string(subj), zap.Error(err))
	}
}

func Publish[Data any](subj constant.Subject, data Data) (isSuccess bool) {
	if natsConn == nil {
		return false
	}
	str, err := json.Marshal(&data)
	if err != nil {
		natsLogger.Error(string(subj), zap.Error(err))
		return
	}
	if err = natsConn.Publish(string(subj), str); err != nil {
		natsLogger.Error(string(subj), zap.Error(err))
		return
	}
	return true
}
