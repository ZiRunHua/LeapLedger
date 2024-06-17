package globalTask

import (
	"KeepAccount/global/task/model"
	"database/sql"
	"errors"
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"time"
)

var cronLogger *zap.Logger
var scheduler *gocron.Scheduler

func NewTransactionCron(handler func(db *gorm.DB) error) func() {
	return func() {
		err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).Transaction(handler)
		if err != nil {
			cronLogger.Error("cronOfPublishRetryTask", zap.Error(err))
		}
	}
}
func cronOfPublishRetryTask(db *gorm.DB) error {
	var retryTask model.RetryTask
	rows, err := db.Model(&retryTask).
		Where("next_exec_time <= ? AND status = ?", time.Now(), model.RetryStatusOfNormal).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Rows()
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			cronLogger.Error("cronOfPublishRetryTask:retryTask.rows.Close", zap.Error(err))
		}
	}(rows)

	for rows.Next() {
		err = db.ScanRows(rows, &retryTask)
		if err != nil {
			cronLogger.Error("execRetryTask:db.ScanRows", zap.Error(err))
			continue
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			return taskServer.republishRetryTask(retryTask, tx)
		})
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
