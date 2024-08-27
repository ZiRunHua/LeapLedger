package transactionService

import (
	"KeepAccount/global/cron"
	"KeepAccount/global/db"
	"KeepAccount/global/nats"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"gorm.io/gorm"
	"time"
)

type _task struct{}

func init() {
	nats.SubscribeTaskWithPayload[transactionModel.StatisticData](
		nats.TaskStatisticUpdate, func(data transactionModel.StatisticData, ctx context.Context) error {
			return GroupApp.Transaction.updateStatistic(data, db.Get(ctx))
		},
	)

	nats.SubscribeTaskWithPayload[transactionModel.Transaction](
		nats.TaskTransactionSync, GroupApp.Transaction.SyncToMappingAccount,
	)
	// timing
	_, err := cron.Scheduler.Every(1).Day().At("00:00").Do(
		cron.PublishTaskWithMakePayload[taskTransactionTimingTaskAssign](
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

	nats.SubscribeTaskWithPayload[taskTransactionTimingTaskAssign](
		nats.TaskTransactionTimingTaskAssign, func(assign taskTransactionTimingTaskAssign, ctx context.Context) error {
			return GroupApp.Timing.Exec.GenerateAndPublishTasks(assign.Deadline, assign.TaskSize, ctx)
		},
	)

	nats.SubscribeTaskWithPayload[transactionTimingExecTask](
		nats.TaskTransactionTimingExec, func(execTask transactionTimingExecTask, ctx context.Context) error {
			return GroupApp.Timing.Exec.ProcessWaitExecByStartId(execTask.StartId, execTask.Size, ctx)
		},
	)
}

func (t *_task) updateStatistic(data transactionModel.StatisticData, tx *gorm.DB) error {
	isSuccess := nats.PublishTaskWithPayload[transactionModel.StatisticData](nats.TaskStatisticUpdate, data)
	if false == isSuccess {
		return server.updateStatistic(data, tx)
	}
	return nil
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
		return nats.ErrNatsNotWork
	}
	return nil
}
