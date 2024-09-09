package transaction

import (
	_ "KeepAccount/test/initialize"
)
import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats"
	accountModel "KeepAccount/model/account"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/test"
	"context"
	"reflect"
	"testing"
	"time"
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
	err = db.Transaction(
		context.TODO(), func(ctx *cus.TxContext) error {
			createOption, err := service.NewOptionFormConfig(transInfo, ctx)
			if err != nil {
				return err
			}
			createOption.WithSyncUpdateStatistic(false)
			trans, err = service.Create(transInfo, user, createOption, ctx)
			return err
		},
	)
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
	defer rows.Close()
	if err != nil {
		t.Error(err)
	}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &transaction)
		if err != nil {
			t.Error(err)
		}
		nats.PublishTaskWithPayload(nats.TaskStatisticUpdate, transaction.GetStatisticData(true))
	}
}
