package transaction

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	categoryModel "github.com/ZiRunHua/LeapLedger/model/category"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	_ "github.com/ZiRunHua/LeapLedger/test/initialize"
)
import (
	"testing"
)

func TestAccount(t *testing.T) {
	t.Parallel()
	user, err := build.User()
	if err != nil {
		t.Fatal(err)
	}
	account, _, err := build.Account(user, accountModel.TypeShare)
	if err != nil {
		t.Fatal(err)
	}
	childUser, err := build.User()
	if err != nil {
		t.Fatal(err)
	}
	childAccount, _, err := build.Account(childUser, accountModel.TypeIndependent)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Share.AddAccountUser(
		account, childUser, accountModel.UserPermissionOwnEditor, context.TODO(),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Share.MappingAccount(childUser, account, childAccount, context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	err = checkMappingAccountTrans(childUser, account, childAccount, constant.Expense)
	if err != nil {
		t.Fatal(err)
	}
}

func checkMappingAccountTrans(
	user userModel.User,
	account, childAccount accountModel.Account,
	ie constant.IncomeExpense) error {
	category, err := query.Category(account.ID, ie)
	if err != nil {
		return err
	}
	childCategory, err := query.Category(childAccount.ID, ie)
	if err != nil {
		return err
	}
	_, err = categoryService.MappingCategory(category, childCategory, user, context.TODO())
	if err != nil {
		return err
	}

	create := func(
		account accountModel.Account,
		user userModel.User,
		category, childCategory categoryModel.Category) error {
		aUser, err := accountModel.NewDao().SelectUser(account.ID, user.ID)
		if err != nil {
			return err
		}
		_, err = CreateMultiTransByCategory(aUser, category, 10)
		if err != nil {
			return err
		}
		return nil
	}
	err = create(account, user, category, childCategory)
	if err != nil {
		return err
	}
	err = create(childAccount, user, childCategory, category)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 60)
	condition := transactionModel.NewStatisticConditionBuilder(category.AccountId)
	condition.WithCategoryIds([]uint{category.ID})
	statistic, err := transactionModel.NewStatisticDao().GetAmountCountByCondition(
		*condition.Build(),
		ie,
	)
	if err != nil {
		return err
	}
	condition = transactionModel.NewStatisticConditionBuilder(childCategory.AccountId)
	condition.WithCategoryIds([]uint{childCategory.ID})
	childStatistic, err := transactionModel.NewStatisticDao().GetAmountCountByCondition(
		*condition.Build(),
		ie,
	)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(statistic, childStatistic) {
		return errors.New(fmt.Sprint("statistic not equal", statistic, childStatistic))
	}
	return nil
}

func CreateMultiTransByCategory(
	aUser accountModel.User, category categoryModel.Category,
	count int,
) (statistic global.IEStatistic, err error) {
	user, err := userModel.NewDao().SelectById(aUser.UserId)
	if err != nil {
		return
	}
	option := transService.NewDefaultOption()
	option.IsSyncTrans()
	for i := 0; i < count; i++ {
		data := build.TransInfo(user, category)
		_, err = transService.Create(data, aUser, transactionModel.RecordTypeOfManual, option, context.TODO())
		if err != nil {
			return
		}
	}
	return
}
