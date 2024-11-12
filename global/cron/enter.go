package cron

import (
	"context"
	"path/filepath"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/global/nats"
	"github.com/ZiRunHua/LeapLedger/initialize"
	"go.uber.org/zap"
)

var (
	logPath = filepath.Join(constant.LogPath, "cron.log")
	logger  *zap.Logger

	Scheduler = initialize.Scheduler
)

func init() {
	var err error
	logger, err = global.Config.Logger.New(logPath)
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
