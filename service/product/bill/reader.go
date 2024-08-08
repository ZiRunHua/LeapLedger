package bill

import (
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	productModel "KeepAccount/model/product"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
)

var logger *zap.Logger

func init() {
	if err := initLogger(); err != nil {
		panic(err)
	}
}

func initLogger() error {
	path := constant.LOG_PATH + "/service/product/bill.log"
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	logFile, err := os.Create(path)
	if err != nil {
		return err
	}
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(logFile),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)
	logger = zap.New(core, zap.AddCaller())
	return nil
}

type ReaderTemplate struct {
	info             productModel.Bill
	ptcMapping       map[constant.IncomeExpense]map[string]productModel.TransactionCategory
	transDataMapping transactionDataColumnMapping
	ptcIdToMapping   map[uint]productModel.TransactionCategoryMapping
	TransactionReadIterator
	SuccessTransList []transactionModel.Transaction
	FailTransList    [][]string
	TransactionReader
	err error
}

type TransactionReadIterator struct {
	currentRow         []string
	currentIndex       int
	currentTransaction transactionModel.Transaction
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

type TransactionReader interface {
	readTransaction() (isSuccess bool)
}

func (r *ReaderTemplate) Init(account *accountModel.Account, product *productModel.Product) error {
	var err error
	r.ptcIdToMapping, err = (&productModel.TransactionCategoryMapping{}).GetPtcIdMapping(account, product.Key)
	if err != nil {
		return err
	}
	r.ptcMapping, err = (&productModel.TransactionCategory{}).GetIncomeExpenseAndNameMap(product.Key)
	if err != nil {
		return err
	}
	r.currentTransaction = transactionModel.Transaction{
		Info: transactionModel.Info{
			AccountId: account.ID,
		},
	}
	bill, err := product.GetBill()
	if err != nil {
		return err
	}
	r.info = *bill
	r.transDataMapping = transactionDataColumnMapping{}
	return nil
}

func (b *ReaderTemplate) ReaderTransFormFile(file *util.FileWithSuffix) (err error) {
	reader := file.GetReaderByEncoding(b.info.Encoding)
	var fileReadAndHandleFunc util.IteratorsHandleReaderFunc
	switch file.Suffix {
	case ".csv":
		fileReadAndHandleFunc = util.File.IteratorsHandleCSVReader
	case ".excel":
		fileReadAndHandleFunc = util.File.IteratorsHandleEXCELReader
	default:
		panic("不支持该文件类型")
	}
	if err = fileReadAndHandleFunc(reader, b.handleRow); err != nil {
		return err
	}
	return
}

func (b *ReaderTemplate) handleRow(row []string, err error) (isContinue bool) {
	b.currentIndex++
	isContinue = true
	if b.currentIndex < b.info.StartRow {
		// 未到读取的起始行 忽略
		return
	} else if b.currentIndex == b.info.StartRow {
		// 处理列标题行
		if b.err = b.setTransDataMapping(row); b.err != nil {
			logger.Error("读取标题行", zap.Strings("data", row), zap.Error(err))
			return false
		}
	} else {
		b.currentRow = row
		if true == b.readTransaction() {
			b.SuccessTransList = append(b.SuccessTransList, b.currentTransaction)
		} else {
			if b.err != nil {
				logger.Error("读取交易行", zap.Strings("data", row), zap.Error(err))
				b.err = nil
			}
			b.FailTransList = append(b.FailTransList, row)
		}
	}
	return
}

// 设置交易数据与列的映射，以确定交易数据所处的列
func (b *ReaderTemplate) setTransDataMapping(header []string) error {
	headerMappedToPtc, err := (&productModel.BillHeader{}).GetNameMap(b.info.ProductKey)
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
	return nil
}
