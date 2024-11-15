package productService

import (
	"context"
	"io"

	"github.com/ZiRunHua/LeapLedger/global/db"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	productModel "github.com/ZiRunHua/LeapLedger/model/product"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	"github.com/ZiRunHua/LeapLedger/service/product/bill"
	"github.com/ZiRunHua/LeapLedger/util/fileTool"
)

type BillFile struct {
	fileName   string
	fileReader io.Reader
}

func (bf *BillFile) GetRowReader() (func(yield func([]string) bool), error) {
	return fileTool.NewRowReader(
		bf.fileReader,
		fileTool.GetFileSuffix(bf.fileName),
	)
}

func (proService *Product) GetNewBillFile(fileName string, fileReader io.Reader) BillFile {
	return BillFile{fileName: fileName, fileReader: fileReader}
}

func (proService *Product) ProcessesBill(
	file BillFile, product productModel.Product, accountUser accountModel.User,
	handler func(transInfo transactionModel.Info, err error) error, ctx context.Context,
) error {
	rowReader, err := file.GetRowReader()
	if err != nil {
		return err
	}
	account, err := accountModel.NewDao(db.Get(ctx)).SelectById(accountUser.AccountId)
	transReader, err := bill.NewReader(account, product, ctx)
	if err != nil {
		return err
	}

	var (
		transInfo transactionModel.Info
		ignore    bool
	)
	for row := range rowReader {
		transInfo, ignore, err = transReader.ReaderTrans(row, ctx)
		if ignore {
			continue
		}
		transInfo.AccountId, transInfo.UserId = accountUser.AccountId, accountUser.UserId
		err = handler(transInfo, err)
		if err != nil {
			return err
		}
	}
	return nil
}
func (proService *Product) BuildTransactionHandler(
	config userModel.BillImportConfig,
	accountUser accountModel.User) func(transactionModel.Info, context.Context) (transactionModel.Transaction, error) {
	option := transactionServer.NewDefaultOption()
	return func(transInfo transactionModel.Info, ctx context.Context) (transactionModel.Transaction, error) {
		trans, err := transactionServer.Create(transInfo, accountUser, transactionModel.RecordTypeOfImport, option, ctx)
		if err == nil {
			return trans, err
		}

		return trans, err
	}
}
