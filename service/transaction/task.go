package transactionService

import (
	"KeepAccount/global/contextKey"
	gTask "KeepAccount/global/task"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"gorm.io/gorm"
	"time"
)

type _task struct{}

func init() {
	gTask.Subscribe[transactionModel.StatisticData](
		gTask.TaskStatisticUpdate, func(data transactionModel.StatisticData, ctx context.Context) error {
			return GroupApp.Transaction.updateStatistic(data, ctx.Value(contextKey.Tx).(*gorm.DB))
		},
	)

	gTask.Subscribe[transactionModel.Transaction](
		gTask.TaskTransactionSync, GroupApp.Transaction.SyncToMappingAccount,
	)
	// timing
	tingEveryCron := func() {
		now := time.Now()
		gTask.Publish[taskTransactionTimingTaskAssign](gTask.TaskTransactionTimingTaskAssign,
			taskTransactionTimingTaskAssign{
				Deadline: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local),
				taskSize: 50,
			})
	}
	_, err := gTask.Scheduler.Every(1).Day().Do(tingEveryCron)
	if err != nil {
		panic(err)
	}

	gTask.Subscribe[taskTransactionTimingTaskAssign](
		gTask.TaskTransactionTimingTaskAssign, func(assign taskTransactionTimingTaskAssign, ctx context.Context) error {
			return GroupApp.Timing.Exec.GenerateAndPublishTasks(assign.Deadline, assign.taskSize, ctx)
		},
	)

	gTask.Subscribe[transactionTimingExecTask](
		gTask.TaskTransactionTimingExec, func(execTask transactionTimingExecTask, ctx context.Context) error {
			return GroupApp.Timing.Exec.ProcessWaitExecByStartId(execTask.StartId, execTask.Size, ctx)
		},
	)
}

func (t *_task) updateStatistic(data transactionModel.StatisticData, tx *gorm.DB) error {
	isSuccess := gTask.Publish[transactionModel.StatisticData](gTask.TaskStatisticUpdate, data)
	if false == isSuccess {
		return server.updateStatistic(data, tx)
	}
	return nil
}

func (t *_task) syncToMappingAccount(trans transactionModel.Transaction, ctx context.Context) error {
	isSuccess := gTask.Publish[transactionModel.Transaction](gTask.TaskTransactionSync, trans)
	if false == isSuccess {
		return server.SyncToMappingAccount(trans, ctx)
	}
	return nil
}

type taskTransactionTimingTaskAssign struct {
	Deadline time.Time
	taskSize int
}

type transactionTimingExecTask struct {
	StartId uint
	Size    int
}

func (t *_task) execTransactionTiming(startId uint, size int) error {
	isSuccess := gTask.Publish[transactionTimingExecTask](gTask.TaskTransactionTimingExec, transactionTimingExecTask{
		StartId: startId, Size: size,
	})
	if !isSuccess {
		return gTask.ErrNatsNotWork
	}
	return nil
}
