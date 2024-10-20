package transaction

import (
	"context"
	"testing"
	"time"

	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/util/timeTool"
)

func TestTiming(t *testing.T) {
	t.Parallel()
	transInfo := get.TransInfo()
	now := time.Now()
	for i := 0; i < 10; i++ {
		transInfo.TradeTime = now.Add(-time.Hour * 24)
		_ = create(
			transactionModel.Timing{
				AccountId:  transInfo.AccountId,
				UserId:     transInfo.UserId,
				TransInfo:  transInfo,
				Type:       transactionModel.Once,
				OffsetDays: 0,
				NextTime:   transInfo.TradeTime,
				Close:      false,
			}, t,
		)

		transInfo.TradeTime = now.Add(-time.Hour * 24)
		_ = create(
			transactionModel.Timing{
				AccountId:  transInfo.AccountId,
				UserId:     transInfo.UserId,
				TransInfo:  transInfo,
				Type:       transactionModel.EveryDay,
				OffsetDays: 0,
				NextTime:   transInfo.TradeTime,
				Close:      false,
			}, t,
		)
		transInfo.TradeTime = now.Add(-time.Hour * 24)
		offsetDays := transInfo.TradeTime.Weekday()
		if offsetDays == 0 {
			offsetDays = 7
		}
		_ = create(
			transactionModel.Timing{
				AccountId:  transInfo.AccountId,
				UserId:     transInfo.UserId,
				TransInfo:  transInfo,
				Type:       transactionModel.EveryWeek,
				OffsetDays: int(offsetDays),
				NextTime:   transInfo.TradeTime,
				Close:      false,
			}, t,
		)

		transInfo.TradeTime = timeTool.GetFirstSecondOfMonth(now)
		_ = create(
			transactionModel.Timing{
				AccountId:  transInfo.AccountId,
				UserId:     transInfo.UserId,
				TransInfo:  transInfo,
				Type:       transactionModel.EveryMonth,
				OffsetDays: transInfo.TradeTime.Day(),
				NextTime:   transInfo.TradeTime,
				Close:      false,
			}, t,
		)
		lastDayOfMonth := timeTool.GetLastSecondOfMonth(timeTool.GetFirstSecondOfMonth(now).AddDate(0, -1, 0))
		transInfo.TradeTime = timeTool.ToDay(lastDayOfMonth)
		_ = create(
			transactionModel.Timing{
				AccountId:  transInfo.AccountId,
				UserId:     transInfo.UserId,
				TransInfo:  transInfo,
				Type:       transactionModel.LastDayOfMonth,
				OffsetDays: transInfo.TradeTime.Day(),
				NextTime:   transInfo.TradeTime,
				Close:      false,
			}, t,
		)
	}
	err := db.Transaction(
		context.TODO(), func(ctx *cus.TxContext) error {
			return service.Timing.Exec.GenerateAndPublishTasks(now, 10, ctx)
		},
	)
	time.Sleep(time.Second * 30)
	if err != nil {
		t.Fatal(err)
	}
	var list []transactionModel.TimingExec
	err = db.Db.Model(&transactionModel.TimingExec{}).Where(
		"account_id = ? AND status != ?", transInfo.AccountId, transactionModel.TimingExecSuccess,
	).Joins("LEFT JOIN transaction_timing ON transaction_timing.id = transaction_timing_exec.config_id").Limit(50).Order("transaction_timing.id DESC").Find(&list).Error
	if err != nil {
		t.Fatal(err)
	}
	if len(list) > 0 {
		t.Fatal("exist exec fail timing", list)
	}
}

func create(testTiming transactionModel.Timing, t *testing.T) transactionModel.Timing {
	ctx := context.TODO()
	var err error
	var timing transactionModel.Timing
	err = db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			timing, err = service.Timing.CreateTiming(testTiming, ctx)
			return err
		},
	)
	if err != nil {
		t.Error(err)
	}
	testTiming.ID = timing.ID
	testTiming.UpdatedAt = timing.UpdatedAt
	testTiming.CreatedAt = timing.CreatedAt
	testTiming.DeletedAt = timing.DeletedAt
	return timing
}
