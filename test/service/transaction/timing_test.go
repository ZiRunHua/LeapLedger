package transaction

import (
	"KeepAccount/global/contextKey"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/test"
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"reflect"
	"testing"
	"time"
)

func TestTiming(t *testing.T) {
	testTiming := test.NewTransTime()
	for i := 0; i < 10; i++ {
		testTiming.Type = transactionModel.Once
		_ = create(testTiming, t)
		testTiming.Type = transactionModel.EveryDay
		_ = create(testTiming, t)
		testTiming.Type = transactionModel.EveryWeek
		_ = create(testTiming, t)
		testTiming.Type = transactionModel.EveryMonth
		_ = create(testTiming, t)
		testTiming.Type = transactionModel.LastDayOfMonth
		_ = create(testTiming, t)
	}
	now := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, 0, 0, 0, 0, time.Local)
	err := db.Transaction(func(tx *gorm.DB) error {
		return service.Timing.Exec.GenerateAndPublishTasks(now, 5, context.WithValue(context.TODO(), contextKey.Tx, tx))
	})
	time.Sleep(time.Second * 10)
	if err != nil {
		t.Error(err)
	}
}

func create(testTiming transactionModel.Timing, t *testing.T) transactionModel.Timing {
	ctx := context.TODO()
	var err error
	var timing transactionModel.Timing
	err = db.Transaction(func(tx *gorm.DB) error {
		timing, err = service.Timing.CreateTiming(testTiming, context.WithValue(ctx, contextKey.Tx, tx))
		return err
	})
	if err != nil {
		t.Error(err)
	}
	testTiming.ID = timing.ID
	testTiming.UpdatedAt = timing.UpdatedAt
	testTiming.CreatedAt = timing.CreatedAt
	testTiming.DeletedAt = timing.DeletedAt
	if !checkTiming(testTiming, timing) {
		t.Error(errors.New("create now equal"))
	}
	return timing
}

func checkTiming(source, target transactionModel.Timing) bool {
	return reflect.DeepEqual(source, target)
}
