package transaction

import (
	"context"

	_ "github.com/ZiRunHua/LeapLedger/test/initialize"
)
import (
	"reflect"
	"testing"
	"time"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/global/nats"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	"github.com/ZiRunHua/LeapLedger/test"
)

func TestCreate(t *testing.T) {
	transInfo := test.NewTransInfo()
	user, err := accountModel.NewDao().SelectUser(transInfo.AccountId, transInfo.UserId)
	if err != nil {
		t.Fatal(err)
	}
	builder := transactionModel.NewStatisticConditionBuilder(transInfo.AccountId)
	builder.WithUserIds([]uint{transInfo.UserId}).WithCategoryIds([]uint{transInfo.CategoryId})
	builder.WithDate(transInfo.TradeTime, transInfo.TradeTime)
	total, err := transactionModel.NewDao().GetIeStatisticByCondition(&transInfo.IncomeExpense, *builder.Build(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var trans transactionModel.Transaction
	createOption, err := service.NewOptionFormConfig(transInfo, context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	trans, err = service.Create(transInfo, user, transactionModel.RecordTypeOfManual, createOption, context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 10)
	newTotal, err := transactionModel.NewDao().GetIeStatisticByCondition(
		&transInfo.IncomeExpense, *builder.Build(), nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if transInfo.IncomeExpense == constant.Income {
		total.Income.Amount += int64(trans.Amount)
		total.Income.Count++
	} else {
		total.Expense.Amount += int64(trans.Amount)
		total.Expense.Count++
	}
	if !reflect.DeepEqual(total, newTotal) {
		t.Fatal("total not equal", total, newTotal)
	} else {
		t.Log("pass", total, newTotal)
	}
}

func TestAll(t *testing.T) {
	var transaction transactionModel.Transaction
	rows, err := global.GvaDb.Model(&transactionModel.Transaction{}).Rows()
	defer func() {
		err = rows.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &transaction)
		if err != nil {
			t.Fatal(err)
		}
		nats.PublishTaskWithPayload(nats.TaskStatisticUpdate, transaction.GetStatisticData(true))
	}
}
