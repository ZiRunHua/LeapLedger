package productService

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	productModel "KeepAccount/model/product"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/service/product/bill"
	"KeepAccount/util/fileTool"
	"context"
	"io"
)

type BillFile struct {
	fileName   string
	fileReader io.Reader
}

func (bf *BillFile) GetRowReader(encoding constant.Encoding) (func(yield func([]string) bool), error) {
	return fileTool.NewRowReader(
		fileTool.GetReaderByEncoding(bf.fileReader, encoding),
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
	billConfig, err := productModel.NewDao(db.Get(ctx)).SelectBillByKey(product.Key)
	if err != nil {
		return err
	}
	rowReader, err := file.GetRowReader(billConfig.Encoding)
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
		err = handler(transInfo, err)
		if err != nil {
			return err
		}
	}
	return nil
}
