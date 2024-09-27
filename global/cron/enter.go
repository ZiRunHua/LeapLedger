package cron

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/nats"
	"KeepAccount/initialize"
	"KeepAccount/util/log"
	"context"
	"go.uber.org/zap"
)

const logPath = constant.LOG_PATH + "/cron.log"

var (
	logger *zap.Logger

	Scheduler = initialize.Scheduler
)

func init() {
	var err error
	logger, err = log.GetNewZapLogger(logPath)
	if err != nil {
		panic(err)
	}
	_, err = Scheduler.Every(30).Minute().Do(
		MakeJobFunc(
			func() error {
				return nats.RepublishDieMsg(50, context.TODO())
			},
		),
	)
	if err != nil {
		panic(err)
	}
}
