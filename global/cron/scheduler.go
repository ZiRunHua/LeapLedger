package cron

import (
	"KeepAccount/initialize"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	cronLogger *zap.Logger

	Scheduler = initialize.Scheduler
)

func cronOfPublishRetryTask(db *gorm.DB) error {

	return nil
}
