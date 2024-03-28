package productService

import (
	accountModel "KeepAccount/model/account"
	productModel "KeepAccount/model/product"
	"KeepAccount/service/product/bill"
	"KeepAccount/util"
	"gorm.io/gorm"
)

type ProductBillImport struct {
	accountUser accountModel.User
	account     accountModel.Account
	product     productModel.Product
	billReader  *bill.ReaderTemplate
}

func newProductBillImport(
	accountUser accountModel.User, account accountModel.Account, product productModel.Product,
) *ProductBillImport {
	var currentBill bill.ReaderTemplate
	//根据第三方产品设置当前账单的读取器
	switch product.Key {
	case productModel.AliPay:
		aliPayReader := &bill.AliPayReader{ReaderTemplate: &currentBill}
		currentBill = bill.ReaderTemplate{TransactionReader: aliPayReader}
	case productModel.WeChatPay:
		weChatPayReader := &bill.WeChatPayReader{ReaderTemplate: &currentBill}
		currentBill = bill.ReaderTemplate{TransactionReader: weChatPayReader}
	default:
		panic("未开放该第三方账本导入功能")
	}
	return &ProductBillImport{
		accountUser: accountUser,
		account:     account,
		product:     product,
		billReader:  &currentBill,
	}
}

func (pbiService *ProductBillImport) init() error {
	var err error
	err = pbiService.billReader.Init(&pbiService.account, &pbiService.product)
	if err != nil {
		return err
	}
	return nil
}

func (pbiService *ProductBillImport) doImport(file *util.FileWithSuffix, tx *gorm.DB) error {
	if err := pbiService.billReader.ReaderTransFormFile(file); err != nil {
		return err
	}
	_, err := transactionServer.CreateMultiple(
		pbiService.accountUser, pbiService.account, pbiService.billReader.SuccessTransList, tx,
	)
	return err
}
