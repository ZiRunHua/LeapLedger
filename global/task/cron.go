package globalTask

import (
	"KeepAccount/global/cusCtx"
	"KeepAccount/global/db"
	"KeepAccount/global/task/model"
	"context"
	"errors"
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"time"
)

var cronLogger *zap.Logger
var Scheduler *gocron.Scheduler

func NewTransactionCron(handler func(db *gorm.DB) error) func() {
	return func() {
		err := db.Transaction(
			context.TODO(), func(ctx *cusCtx.TxContext) error {
				tx := ctx.GetDb()
				return handler(tx)
			},
		)
		if err != nil {
			cronLogger.Error("cronOfPublishRetryTask", zap.Error(err))
		}
	}
}

func cronOfPublishRetryTask(db *gorm.DB) error {
	db = db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)})
	var list []model.RetryTask
	err := db.Where("next_exec_time <= ? AND status = ?", time.Now(), model.RetryStatusOfNormal).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Find(&list).Error
	if err != nil {
		return err
	}
	for _, retryTask := range list {
		err = db.Transaction(
			func(tx *gorm.DB) error {
				return taskServer.republishRetryTask(retryTask, tx)
			},
		)
		if err != nil {
			if errors.Is(err, ErrRepublishFailure) {
				continue
			} else {
				// If the error is not a publishing failure, the retryTask is set as an exception to avoid repeated publishing
				err = retryTask.Abnormal(db)
				if err != nil {
					cronLogger.Error("execRetryTask:retryTask.Abnormal", zap.Error(err))
				}
				continue
			}
		}
	}
	return nil
}
