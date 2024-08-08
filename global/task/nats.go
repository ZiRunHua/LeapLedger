package globalTask

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/contextKey"
	"KeepAccount/global/task/model"
	"context"
	"encoding/json"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
	"time"
)

var (
	config = global.Config.Nats

	natsConn   *nats.Conn
	natsLogger *zap.Logger
	natsServer *server.Server

	ErrNatsNotWork = errors.New("nats not work")
)

// user task
const TaskCreateTourist constant.Subject = "createTourist"

// transaction task
const TaskStatisticUpdate constant.Subject = "statisticUpdate"
const TaskTransactionSync constant.Subject = "transactionSync"
const TaskTransactionTimingExec constant.Subject = "transactionTimingExec"
const TaskTransactionTimingTaskAssign constant.Subject = "transactionTimingTaskAssign"

// category task
const TaskMappingCategoryToAccountMapping constant.Subject = "mappingCategoryToAccountMapping"
const TaskUpdateCategoryMapping constant.Subject = "updateCategoryMapping"

func Subscribe[DataType any](subj constant.Subject, handleFunc transactionHandle[DataType]) {
	subscribe[DataType](subj, handleFunc)
}

func subscribe[DataType any](subject constant.Subject, handleTransaction transactionHandle[DataType]) {
	if natsConn == nil || !config.CanSubscribe(subject) {
		return
	}

	executeTransaction := func(msgData []byte, dbTransaction *gorm.DB) error {
		var data DataType
		if err := json.Unmarshal(msgData, &data); err != nil {
			return err
		}
		return dbTransaction.Transaction(func(tx *gorm.DB) error {
			return handleTransaction(data, context.WithValue(context.Background(), contextKey.Tx, tx))
		})
	}

	msgHandler := func(msg *nats.Msg) {
		if isRetryTask(msg) {
			taskId, err := strconv.ParseUint(string(msg.Data), 10, 0)
			if err != nil {
				natsLogger.Error(string(subject), zap.Error(errors.WithMessage(err, "task retry:task_id")))
				return
			}
			retryTransaction := func(tx *gorm.DB) error {
				var task model.Task
				err = db.First(&task, taskId).Error
				if err != nil {
					return errors.WithMessage(err, "select task")
				}
				err = executeTransaction(msg.Data, tx)
				if err != nil {
					return err
				}
				err = task.Complete(tx)
				if err != nil {
					return err
				}
				return nil
			}
			err = db.Transaction(retryTransaction)
			if err != nil {
				err = db.Model(&model.Task{}).Where("id = ?", taskId).Update("error", err.Error()).Error
				if err != nil {
					natsLogger.Error(string(subject), zap.Error(errors.WithMessage(err, "task error update")))
					return
				}
			}
		} else {
			err := executeTransaction(msg.Data, db)
			if err != nil {
				msgProcessFail(msg, err)
				return
			}
		}
	}

	_, err := natsConn.Subscribe(string(subject), msgHandler)
	if err != nil {
		natsLogger.Error(string(subject), zap.Error(err))
	}
}

func isRetryTask(msg *nats.Msg) bool { return isDigits(msg.Data) }

func msgProcessFail(msg *nats.Msg, execErr error) {
	task, err := taskServer.addFailedTask(constant.Subject(msg.Subject), string(msg.Data), execErr)
	if err != nil {
		natsLogger.Error("taskServer.addFailedTask", zap.Error(err))
		return
	}
	_, err = taskServer.addRetryTask(task)
	if err != nil {
		natsLogger.Error("taskServer.addRetryTask", zap.Error(err))
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
	return publish(subj, str)
}

func publish(subj constant.Subject, data []byte) bool {
	err := natsConn.Publish(string(subj), data)
	if err != nil {
		natsLogger.Error(string(subj), zap.Error(err))
		return false
	}
	return true
}

func publishTask(subj constant.Subject, task model.Task) bool {
	err := natsConn.Publish(string(subj), strconv.AppendUint([]byte{}, uint64(task.ID), 10))
	if err != nil {
		natsLogger.Error(string(subj), zap.Error(err))
		return false
	}
	return true
}

func backOff(count uint8) time.Duration {
	switch count {
	case 0:
		return time.Second * 3
	case 1:
		return time.Second * 30
	case 2:
		return time.Minute * 5
	case 3:
		return time.Minute * 50
	case 4:
		return time.Minute * 500
	default:
		return time.Minute * 5000
	}
}

func isDigits(b []byte) bool {
	for _, v := range b {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}
