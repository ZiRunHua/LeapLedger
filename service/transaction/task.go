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
	nats.Subscribe[transactionModel.StatisticData](
		nats.TaskStatisticUpdate, func(data transactionModel.StatisticData, ctx context.Context) error {
			return GroupApp.Transaction.updateStatistic(data, db.Get(ctx))
		},
	)

	nats.Subscribe[transactionModel.Transaction](
		nats.TaskTransactionSync, GroupApp.Transaction.SyncToMappingAccount,
	)
	// timing
	tingEveryCron := func() {
		now := time.Now()
		nats.Publish[taskTransactionTimingTaskAssign](nats.TaskTransactionTimingTaskAssign,
			taskTransactionTimingTaskAssign{
				Deadline: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local),
				TaskSize: 50,
			})
	}
	_, err := cron.Scheduler.Every(1).Day().At("00:00").Do(tingEveryCron)
	if err != nil {
		panic(err)
	}

	nats.Subscribe[taskTransactionTimingTaskAssign](
		nats.TaskTransactionTimingTaskAssign, func(assign taskTransactionTimingTaskAssign, ctx context.Context) error {
			return GroupApp.Timing.Exec.GenerateAndPublishTasks(assign.Deadline, assign.TaskSize, ctx)
		},
	)

	nats.Subscribe[transactionTimingExecTask](
		nats.TaskTransactionTimingExec, func(execTask transactionTimingExecTask, ctx context.Context) error {
			return GroupApp.Timing.Exec.ProcessWaitExecByStartId(execTask.StartId, execTask.Size, ctx)
		},
	)
}

func (t *_task) updateStatistic(data transactionModel.StatisticData, tx *gorm.DB) error {
	isSuccess := nats.Publish[transactionModel.StatisticData](nats.TaskStatisticUpdate, data)
	if false == isSuccess {
		return server.updateStatistic(data, tx)
	}
	return nil
}

func (t *_task) syncToMappingAccount(trans transactionModel.Transaction, ctx context.Context) error {
	isSuccess := nats.Publish[transactionModel.Transaction](nats.TaskTransactionSync, trans)
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
	isSuccess := nats.Publish[transactionTimingExecTask](nats.TaskTransactionTimingExec, transactionTimingExecTask{
		StartId: startId, Size: size,
	})
	if !isSuccess {
		return nats.ErrNatsNotWork
	}
	return nil
}
