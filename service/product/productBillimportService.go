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

func (bf *BillFile) GetRowChan(encoding constant.Encoding) (chan []string, error) {
	return fileTool.NewRowChan(
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
	rowChan, err := file.GetRowChan(billConfig.Encoding)
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
	for row := range rowChan {
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
