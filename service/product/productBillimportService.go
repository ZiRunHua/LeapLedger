package productService

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	productModel "KeepAccount/model/product"
	transactionModel "KeepAccount/model/transaction"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"

	"KeepAccount/util"
	"mime/multipart"
)

type ProductBillImport struct {
	account accountModel.Account
	product productModel.Product
	bill    bill
}

type BillImport interface {
	BillImport(product *productModel.Product, file *multipart.FileHeader) error
}

func (pbiService *ProductBillImport) BillImport(
	account *accountModel.Account, product *productModel.Product, file *multipart.FileHeader, tx *gorm.DB,
) error {
	err := pbiService.init(account, product, file)
	if err != nil {
		return err
	}

	fmt.Println(pbiService.bill.ptcMapping)
	fmt.Println("pbiService.bill.ptcIdToMapping")
	fmt.Println(pbiService.bill.ptcIdToMapping)
	err = pbiService.doImport(tx)
	if err != nil {
		return err
	}
	return nil
}

func (pbiService *ProductBillImport) init(
	account *accountModel.Account, product *productModel.Product, file *multipart.FileHeader,
) error {
	pbiService.account = *account
	pbiService.product = *product
	billInfo, err := product.GetBill()
	if err != nil {
		return err
	}
	BillTransCateMapping, err := (&productModel.TransactionCategoryMapping{}).GetPtcIdMapping(account, product.Key)
	if err != nil {
		return err
	}
	BillPtcMapping, err := (&productModel.TransactionCategory{}).GetIncomeExpenseAndNameMap(product.Key)
	if err != nil {
		return err
	}
	pbiService.bill = bill{
		info: *billInfo,
		currentTransaction: transactionModel.Transaction{
			AccountID: account.ID,
		},
		ptcMapping:       BillPtcMapping,
		ptcIdToMapping:   BillTransCateMapping,
		rows:             [][]string{},
		transDataMapping: transactionDataColumnMapping{},
	}
	return pbiService.bill.handleFile(file)
}

func (pbiService *ProductBillImport) doImport(tx *gorm.DB) error {
	transactionList := pbiService.bill.getTransactionList()
	fmt.Println(transactionList)
	if len(transactionList) == 0 {
		return nil
	}
	_, err := transactionServer.CreateMultiple(&pbiService.account, transactionList, tx)
	return err
}

type bill struct {
	info               productModel.Bill
	ptcMapping         map[global.IncomeExpense]map[string]productModel.TransactionCategory
	transDataMapping   transactionDataColumnMapping
	rows               [][]string
	ptcIdToMapping     map[uint]productModel.TransactionCategoryMapping
	currentRow         []string
	currentTransaction transactionModel.Transaction
	currentIndex       int
	err                error
}

func (b *bill) handleFile(file *multipart.FileHeader) (err error) {
	reader := util.File.GetFileReader(file, b.info.Encoding)
	fileType := util.File.GetFileSuffix(file.Filename)
	switch fileType {
	case ".csv":
		b.rows, err = util.File.GetContentFormCSVReader(reader)
	case ".excel":
		b.rows, err = util.File.GetContentFormEXCELReader(reader)
	default:
		err = errors.Wrap(errors.New("不支持该文件类型"), "Product Service:BillImport")
	}
	if b.info.StartRow > len(b.rows) {
		b.rows = [][]string{}
	} else if b.info.StartRow > 0 {
		err = b.setTransDataMapping(b.rows[b.info.StartRow-1])
		if err != nil {
			return err
		}
		b.rows = (b.rows)[b.info.StartRow:]
	}
	return
}

func (b *bill) setTransDataMapping(header []string) error {
	headerMappedToPtc, err := (&productModel.BillHeader{}).GetNameMap(b.info.ProductKey)
	fmt.Println(headerMappedToPtc)
	if err != nil {
		return err
	}
	headerTypeMappedToColumn := map[productModel.BillHeaderType]int{}
	for index, name := range header {
		name = strings.TrimSpace(name)
		if _, exist := headerMappedToPtc[name]; exist == true {
			headerTypeMappedToColumn[headerMappedToPtc[name].Type] = index
		}
	}
	fmt.Println(headerTypeMappedToColumn)
	needHeader := []productModel.BillHeaderType{
		productModel.TransCategory, productModel.IncomeExpense, productModel.Amount, productModel.Remark,
		productModel.TransTime, productModel.OrderNumber, productModel.TransStatus,
	}
	for i := range needHeader {
		if _, exist := headerTypeMappedToColumn[needHeader[i]]; exist == false {
			return errors.Wrap(errors.New(string(needHeader[i]+"数据缺失")), "setTransMapping")
		}
	}
	b.transDataMapping = transactionDataColumnMapping{
		OrderNumber:   headerTypeMappedToColumn[productModel.OrderNumber],
		TransCategory: headerTypeMappedToColumn[productModel.TransCategory],
		IncomeExpense: headerTypeMappedToColumn[productModel.IncomeExpense],
		Amount:        headerTypeMappedToColumn[productModel.Amount],
		Remark:        headerTypeMappedToColumn[productModel.Remark],
		TradeTime:     headerTypeMappedToColumn[productModel.TransTime],
		TransStatus:   headerTypeMappedToColumn[productModel.TransStatus],
	}
	fmt.Println(b.transDataMapping)
	return nil
}

func (b *bill) getTransactionList() (transactionList []transactionModel.Transaction) {
	if b.rows == nil || len(b.rows) == 0 {
		return
	}
	for index := 1; index < len(b.rows); index++ {
		b.currentRow = b.rows[index]
		//fmt.Println(b.checkOrderStatus())
		//fmt.Println(b.setTransCategory())
		//fmt.Println(b.setAmount())
		if !b.checkOrderStatus() || !b.setTransCategory() || !b.setAmount() || !b.setTradeTime() {
			continue
		}
		b.setRemark()
		transactionList = append(transactionList, b.currentTransaction)
	}
	return
}

type transactionDataColumnMapping struct {
	OrderNumber   int
	TransCategory int
	IncomeExpense int
	Amount        int
	Remark        int
	TradeTime     int
	TransStatus   int
}

func (b *bill) checkOrderStatus() bool {
	status := strings.TrimSpace(b.currentRow[b.transDataMapping.TransStatus])
	fmt.Println(status)
	if status != "交易成功" {
		return false
	}
	return true
}

func (b *bill) setTransCategory() bool {
	incomeExpenseStr := strings.TrimSpace(b.currentRow[b.transDataMapping.IncomeExpense])
	var incomeExpense global.IncomeExpense
	if incomeExpenseStr == "收入" {
		incomeExpense = global.Income
	} else if incomeExpenseStr == "支出" {
		incomeExpense = global.Expense
	} else {
		return false
	}
	name := strings.TrimSpace(b.currentRow[b.transDataMapping.TransCategory])
	ptc, exist := b.ptcMapping[incomeExpense][name]
	if exist == false {
		return false
	}
	mapping, exist := b.ptcIdToMapping[ptc.ID]
	if exist == false {
		return false
	}
	b.currentTransaction.IncomeExpense = incomeExpense
	b.currentTransaction.CategoryID = mapping.CategoryID
	return true
}

func (b *bill) setAmount() bool {
	var amountFloat float64
	amountFloat, b.err = strconv.ParseFloat(b.currentRow[b.transDataMapping.Amount], 64)
	if b.err != nil {
		return false
	} else {
		b.currentTransaction.Amount = int(amountFloat) * 100
	}
	return true
}

func (b *bill) setRemark() {
	b.currentTransaction.Remark = strings.TrimSpace(b.currentRow[b.transDataMapping.Remark])
}

func (b *bill) setTradeTime() bool {
	date := b.currentRow[b.transDataMapping.TradeTime]
	if b.currentTransaction.TradeTime, b.err = time.Parse(b.info.DateFormat, date); b.err != nil {
		return false
	}
	return true
}
