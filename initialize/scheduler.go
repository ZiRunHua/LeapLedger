package initialize

import (
	"github.com/go-co-op/gocron"
	"time"
)

type _scheduler struct {
}

func (m *_scheduler) do() error {
	Scheduler = gocron.NewScheduler(time.Local)
	Scheduler.StartAsync()
	return nil
}
