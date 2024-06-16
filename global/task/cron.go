package globalTask

import (
	"KeepAccount/global/task/model"
	"KeepAccount/initialize"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var cronLogger = initialize.CronLogger
var scheduler = initialize.Scheduler

type retryTasker[Data any] struct {
	Data   Data
	TaskId uint
}

func cronOfPublishRetryTask(db *gorm.DB) error {
	var retryTask model.RetryTask
	rows, err := db.Model(&retryTask).
		Where("next_exec_time <= ? AND status = ?", time.Now(), model.RetryStatusOfNormal).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = db.ScanRows(rows, &retryTask)
		if err != nil {
			cronLogger.Error("execRetryTask:db.ScanRows", zap.Error(err))
			continue
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			return taskServer.republishTask(retryTask, tx)
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
