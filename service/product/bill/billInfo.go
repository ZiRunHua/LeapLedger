package bill

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	productModel "KeepAccount/model/product"
	"context"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type BillInfo struct {
	info             productModel.Bill
	location         *time.Location
	ptcMapping       map[constant.IncomeExpense]map[string]productModel.TransactionCategory
	transDataMapping transactionDataColumnMapping
	ptcIdToMapping   map[uint]productModel.TransactionCategoryMapping
}

func (b *BillInfo) init(bill productModel.Bill, account accountModel.Account, ctx context.Context) error {
	b.info, b.location = bill, account.GetTimeLocation()
	dao := productModel.NewDao(db.Get(ctx))
	var err error
	b.ptcIdToMapping, err = dao.GetPtcIdMapping(account.ID, bill.ProductKey)
	if err != nil {
		return err
	}
	b.ptcMapping, err = dao.GetIncomeExpenseAndNameMap(bill.ProductKey)
	return err
}

func (b *BillInfo) setTransDataMapping(header []string, ctx context.Context) error {
	headerMappedToPtc, err := productModel.NewDao(db.Get(ctx)).GetNameMap(b.info.ProductKey)
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
