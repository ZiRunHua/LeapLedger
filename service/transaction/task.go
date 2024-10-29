package transactionService

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ZiRunHua/LeapLedger/global/cron"
	"github.com/ZiRunHua/LeapLedger/global/nats"
)

type _task struct{}

func init() {
	// update statistic
	nats.SubscribeTaskWithPayload(nats.TaskStatisticUpdate, GroupApp.Transaction.updateStatistic)
	// sync trans
	nats.SubscribeTaskWithPayloadAndProcessInTransaction(
		nats.TaskTransactionSync, GroupApp.Transaction.SyncToMappingAccount,
	)
	// timing
	var moments []string
	for i := 0; i < 24; i++ {
		moments = append(moments, fmt.Sprintf("%02d:00", i))
	}
	_, err := cron.Scheduler.Every(1).Day().At(strings.Join(moments, ";")).Do(
		cron.PublishTaskWithMakePayload(
			nats.TaskTransactionTimingTaskAssign, func() (taskTransactionTimingTaskAssign, error) {
				now := time.Now()
				return taskTransactionTimingTaskAssign{
					Deadline: time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.Local),
					TaskSize: 50,
				}, nil
			},
		),
	)
	if err != nil {
		panic(err)
	}

	nats.SubscribeTaskWithPayloadAndProcessInTransaction(
		nats.TaskTransactionTimingTaskAssign, func(assign taskTransactionTimingTaskAssign, ctx context.Context) error {
			return GroupApp.Timing.Exec.GenerateAndPublishTasks(assign.Deadline, assign.TaskSize, ctx)
		},
	)

	nats.SubscribeTaskWithPayloadAndProcessInTransaction(
		nats.TaskTransactionTimingExec, func(execTask transactionTimingExecTask, ctx context.Context) error {
			return GroupApp.Timing.Exec.ProcessWaitExecByStartId(execTask.StartId, execTask.Size, ctx)
		},
	)
}

type taskTransactionTimingTaskAssign struct {
	Deadline time.Time
	TaskSize int
}

type transactionTimingExecTask struct {
	StartId uint
	Size    int
}
