package transactionService

import (
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/lock"
	accountModel "KeepAccount/model/account"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"go.uber.org/zap"
	"time"
)

type Timing struct {
	Exec TimingExec
}

func (tService *Timing) CreateTiming(timing transactionModel.Timing, ctx context.Context) (
	transactionModel.Timing, error,
) {
	tx := db.Get(ctx)
	if err := timing.TransInfo.Check(tx); err != nil {
		return timing, err
	}
	timing.TransInfo.TradeTime = timing.NextTime
	err := tx.Create(&timing).Error
	return timing, err
}

func (tService *Timing) UpdateTiming(timing transactionModel.Timing, ctx context.Context) (
	transactionModel.Timing, error,
) {
	tx := db.Get(ctx)
	if err := timing.TransInfo.Check(tx); err != nil {
		return timing, err
	}
	timing.TransInfo.TradeTime = timing.NextTime
	err := tx.Where("id = ?", timing.ID).Updates(&timing).Error
	return timing, err
}

type TimingExec struct {
}

func (te *TimingExec) getLock() lock.Lock {
	return lock.NewWithDuration("transaction_timing_exec", time.Minute*10)
}

func (te *TimingExec) execAfterLock(exec func() error, ctx context.Context) (err error) {
	l := te.getLock()
	err = l.Lock(ctx)
	if err != nil {
		return err
	}
	defer func(l lock.Lock, ctx context.Context) {
		if err != nil {
			_ = l.Release(ctx)
		} else {
			err = l.Release(ctx)
		}
	}(l, ctx)
	return exec()
}

func (te *TimingExec) GenerateAndPublishTasks(deadline time.Time, taskSize int, ctx context.Context) error {
	var startIds []uint
	var err error
	exec := func() error {
		startIds, err = te.makeAndSplitExecTask(deadline, taskSize, ctx)
		return err
	}
	err = te.execAfterLock(exec, ctx)
	if err != nil {
		return err
	}
	for _, startId := range startIds {
		err = db.AddCommitCallback(
			ctx, func() {
				var err error
				err = task.execTransactionTiming(startId, taskSize)
				if err != nil {
					errorLog.Error("GenerateAndPublishTasks => execTransactionTiming", zap.Error(err))
				}
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (te *TimingExec) makeAndSplitExecTask(deadline time.Time, size int, ctx context.Context) (
	starIds []uint, err error,
) {
	var (
		tx       = db.Get(ctx)
		count    int
		timeExec transactionModel.TimingExec
	)
	process := func(timing transactionModel.Timing) error {
		timeExec, err = timing.MakeExecTask(tx)
		if err != nil {
			return err
		}
		err = timing.UpdateNextTime(tx)
		if err != nil {
			return err
		}
		if count%size == 0 {
			starIds = append(starIds, timeExec.ID)
		}
		count++
		return nil
	}
	err = transactionModel.NewDao().SelectAllTimingAndProcess(deadline, process)
	return
}

func (te *TimingExec) ProcessWaitExecByStartId(startId uint, limit int, ctx context.Context) error {
	var (
		accountUser accountModel.User
		trans       transactionModel.Transaction
		tx          = db.Get(ctx)
		transDao    = transactionModel.NewDao(tx)
	)
	list, err := transDao.SelectWaitTimingExec(startId, limit)
	if err != nil {
		return err
	}
	for _, timingExec := range list {
		err = db.Transaction(
			ctx, func(ctx *cus.TxContext) error {
				trans, err = server.Create(timingExec.TransInfo, accountUser, server.NewDefaultOption(), ctx)
				return err
			},
		)
		if err != nil {
			err = timingExec.ExecFail(err, tx)
			if err != nil {
				return err
			}
		}
		err = timingExec.ExecSuccess(trans, tx)
		if err != nil {
			return err
		}
	}
	return nil
}
