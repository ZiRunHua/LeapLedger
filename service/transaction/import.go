package transactionService

import (
	"context"

	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
)

type Import struct {
}

func (i *Import) BuildTransactionHandler(
	config userModel.BillImportConfig,
	accountUser accountModel.User) func(context.Context) (transactionModel.Transaction, error) {
	option := server.NewDefaultOption()
	return func(transInfo transactionModel.Info, ctx context.Context) (transactionModel.Transaction, error) {
		trans, err := server.Create(transInfo, accountUser, transactionModel.RecordTypeOfImport, option, ctx)
		if err == nil {
			return trans, err
		}

		return trans, err
	}
}
