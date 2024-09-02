package transaction

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/cusCtx"
	"KeepAccount/global/db"
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
		t.Error(err)
	}
	builder := transactionModel.NewStatisticConditionBuilder(transInfo.AccountId)
	builder.WithUserIds([]uint{transInfo.UserId}).WithCategoryIds([]uint{transInfo.CategoryId})
	builder.WithDate(transInfo.TradeTime, transInfo.TradeTime)
	total, err := transactionModel.NewDao().GetIeStatisticByCondition(&transInfo.IncomeExpense, *builder.Build(), nil)
	if err != nil {
		t.Error(err)
	}
	var trans transactionModel.Transaction
	err = db.Transaction(
		context.TODO(), func(ctx *cusCtx.TxContext) error {
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
		t.Error(err)
	}
	time.Sleep(time.Second * 10)
	newTotal, err := transactionModel.NewDao().GetIeStatisticByCondition(
		&transInfo.IncomeExpense, *builder.Build(), nil,
	)
	if err != nil {
		t.Error(err)
	}

	if transInfo.IncomeExpense == constant.Income {
		total.Income.Amount += int64(trans.Amount)
		total.Income.Count++
	} else {
		total.Expense.Amount += int64(trans.Amount)
		total.Expense.Count++
	}
	if !reflect.DeepEqual(total, newTotal) {
		t.Error("total not equal", total, newTotal)
	} else {
		t.Log("pass", total, newTotal)
	}
}
