// Package bill is used to define the bill reading method
//
// The bill package uses the template method design pattern, so to add a new bill reader,
// all you need to do is implement [TransactionReader] and update the [NewReader] method.
//
// Of course, the new product configuration needs to be completed before this,
// Its configuration is very simple, refer to other product in A file [productModel.initSqlFile]
// set the data of the new product, and set [productModel.Key]
package bill

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/global/db"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	productModel "github.com/ZiRunHua/LeapLedger/model/product"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	logPath = filepath.Clean(constant.LogPath + "/service/product/bill.log")
	logger  *zap.Logger
)

func init() {
	var err error
	if logger, err = global.Config.Logger.New(logPath); err != nil {
		panic(err)
	}
}

func NewReader(account accountModel.Account, product productModel.Product, ctx context.Context) (Reader, error) {
	var reader ReaderTemplate
	switch product.Key {
	case productModel.AliPay:
		reader.TransactionReader = &AliPayReader{}
	case productModel.WeChatPay:
		reader.TransactionReader = &WeChatPayReader{}
	}
	return &reader, reader.init(account, product, ctx)
}

type Reader interface {
	TransactionReader
	ReaderTrans(row []string, ctx context.Context) (trans transactionModel.Info, ignore bool, err error)
	init(account accountModel.Account, product productModel.Product, ctx context.Context) error
}
type ReaderTemplate struct {
	BillInfo

	TransactionReadIterator
	TransactionReader
}

type TransactionReadIterator struct {
	currentRow         []string
	currentIndex       int
	currentTransaction transactionModel.Info
	err                error
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
	readTransaction(*ReaderTemplate) (ignore bool, err error)
}

func (t *ReaderTemplate) init(account accountModel.Account, product productModel.Product, ctx context.Context) error {
	t.currentTransaction = transactionModel.Info{AccountId: account.ID}
	bill, err := productModel.NewDao(db.Get(ctx)).SelectBillByKey(product.Key)
	if err != nil {
		return err
	}
	err = t.BillInfo.init(bill, account, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (t *ReaderTemplate) ReaderTrans(row []string, ctx context.Context) (
	trans transactionModel.Info, ignore bool, err error,
) {
	t.currentIndex++
	ignore = true

	switch t.BillInfo.status {
	case statusOfReadInHead:
		// Try to read the head
		if strings.TrimSpace(row[0]) != t.BillInfo.billHeaders[0].Name {
			return
		}
		if t.err = t.setTransDataMapping(row, ctx); t.err != nil {
			logger.Error("读取标题行", zap.Strings("data", row), zap.Error(err))
			return trans, false, errors.New("读取标题行失败")
		}
		t.BillInfo.status = statusOfReadInTransaction
	case statusOfReadInTransaction:
		t.currentRow = row
		ignore, err = t.readTransaction(t)
		if ignore {
			return trans, ignore, nil
		}
		if err != nil {
			return t.currentTransaction, false, err
		}
		return t.currentTransaction, false, nil
	default:
		panic("error bill status")
	}
	return
}
