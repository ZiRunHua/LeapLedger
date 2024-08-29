package transactionService

import (
	"KeepAccount/global/cron"
	"KeepAccount/global/nats"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"github.com/pkg/errors"
	"time"
)

type _task struct{}

func init() {
	//update statistic
	nats.SubscribeTaskWithPayload(nats.TaskStatisticUpdate, GroupApp.Transaction.updateStatistic)

	//sync trans
	nats.SubscribeTaskWithPayload(nats.TaskTransactionSync, GroupApp.Transaction.SyncToMappingAccount)
	// timing
	_, err := cron.Scheduler.Every(1).Day().At("00:00").Do(
		cron.PublishTaskWithMakePayload(
			nats.TaskTransactionTimingTaskAssign, func() (taskTransactionTimingTaskAssign, error) {
				now := time.Now()
				return taskTransactionTimingTaskAssign{
					Deadline: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local),
					TaskSize: 50,
				}, nil
			},
		),
	)
	if err != nil {
		panic(err)
	}

	nats.SubscribeTaskWithPayload(
		nats.TaskTransactionTimingTaskAssign, func(assign taskTransactionTimingTaskAssign, ctx context.Context) error {
			return GroupApp.Timing.Exec.GenerateAndPublishTasks(assign.Deadline, assign.TaskSize, ctx)
		},
	)

	nats.SubscribeTaskWithPayload(
		nats.TaskTransactionTimingExec, func(execTask transactionTimingExecTask, ctx context.Context) error {
			return GroupApp.Timing.Exec.ProcessWaitExecByStartId(execTask.StartId, execTask.Size, ctx)
		},
	)
}

func (t *_task) syncToMappingAccount(trans transactionModel.Transaction, ctx context.Context) error {
	isSuccess := nats.PublishTaskWithPayload[transactionModel.Transaction](nats.TaskTransactionSync, trans)
	if false == isSuccess {
		return server.SyncToMappingAccount(trans, ctx)
	}
	return nil
}

type taskTransactionTimingTaskAssign struct {
	Deadline time.Time
	TaskSize int
}

type transactionTimingExecTask struct {
	StartId uint
	Size    int
}

func (t *_task) execTransactionTiming(startId uint, size int) error {
	isSuccess := nats.PublishTaskWithPayload[transactionTimingExecTask](
		nats.TaskTransactionTimingExec, transactionTimingExecTask{
			StartId: startId, Size: size,
		},
	)
	if !isSuccess {
		return errors.New("nats not work")
	}
	return nil
}
