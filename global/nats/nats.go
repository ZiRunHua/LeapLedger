package nats

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/initialize"
	"KeepAccount/util/log"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var natsConn = initialize.Nast
var natsLog *zap.Logger

const TaskStatisticUpdate = "statisticUpdate"

func init() {
	var err error
	if natsLog, err = log.GetNewZapLogger(constant.LOG_PAYH + "/nats.log"); err != nil {
		panic(err)
	}
}
func TransSubscribe[Data any](subj string, handleFunc func(*gorm.DB, Data) error) {
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

func Subscribe[T any](subj string, handleFunc func(T) error) {
	if natsConn == nil {
		return
	}
	_, _ = natsConn.Subscribe(
		subj, func(msg *nats.Msg) {
			var t T
			if err := json.Unmarshal(msg.Data, &t); err != nil {
				natsLog.Error(msg.Subject, zap.Error(err))
				return
			}
			err := handleFunc(t)
			if err != nil {
				natsLog.Error(msg.Subject, zap.Error(err))
			}
		},
	)
}

func Publish[Data any](subj string, data Data) (isSuccess bool) {
	if natsConn == nil {
		return false
	}
	str, err := json.Marshal(&data)
	if err != nil {
		natsLog.Error(subj, zap.Error(err))
		return
	}
	if err = natsConn.Publish(subj, str); err != nil {
		natsLog.Error(subj, zap.Error(err))
		return
	}
	return true
}
