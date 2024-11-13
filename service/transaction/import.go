package transactionService

import (
	"context"

	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	productModel "github.com/ZiRunHua/LeapLedger/model/product"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
)

type Import struct {
}

func (i *Import) BuildTransactionHandler(bill productModel.Bill, accountUser accountModel.User) func(transactionModel.Info, context.Context) (transactionModel.Transaction, error) {
	option := server.NewDefaultOption()
	return func(transInfo transactionModel.Info, ctx context.Context) (transactionModel.Transaction, error) {
		trans, err := server.Create(transInfo, accountUser, transactionModel.RecordTypeOfImport, option, ctx)
		if err != nil {
			return trans, err
		}
		// Todo: import config
		return trans, err
	}
}
