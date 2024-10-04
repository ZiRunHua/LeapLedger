package bill

import (
	"KeepAccount/global/constant"
	transactionModel "KeepAccount/model/transaction"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type AliPayReader struct {
	aliPayTransactionReader
}

type aliPayTransactionReader interface {
	TransactionReader
	checkOrderStatus() bool
	setTransCategory() error
	setAmount() error
	setRemark()
	setTradeTime() error
}

func (r *AliPayReader) readTransaction(t *ReaderTemplate) (bool, error) {
	t.currentTransaction = transactionModel.Info{}
	if !r.checkOrderStatus(t) {
		return true, nil
	}
	r.setRemark(t)
	var execErr error
	err := r.setTransCategory(t)
	if err != nil {
		if errors.Is(err, ErrCategoryCannotRead) {
			return true, nil
		}
		execErr = err
	}
	err = r.setAmount(t)
	if err != nil {
		execErr = err
	}
	err = r.setTradeTime(t)
	if err != nil {
		execErr = errors.New("读取交易时间错误：" + err.Error())
	}
	return false, execErr
}

func (r *AliPayReader) checkOrderStatus(t *ReaderTemplate) bool {
	status := strings.TrimSpace(t.currentRow[t.transDataMapping.TransStatus])
	if status != "交易成功" {
		return false
	}
	return true
}

func (r *AliPayReader) setTransCategory(t *ReaderTemplate) error {
	incomeExpenseStr := strings.TrimSpace(t.currentRow[t.transDataMapping.IncomeExpense])
	var incomeExpense constant.IncomeExpense
	if incomeExpenseStr == "收入" {
		incomeExpense = constant.Income
	} else if incomeExpenseStr == "支出" {
		incomeExpense = constant.Expense
	} else {
		return ErrCategoryCannotRead
	}
	name := strings.TrimSpace(t.currentRow[t.transDataMapping.TransCategory])
	ptc, exist := t.ptcMapping[incomeExpense][name]
	if exist == false {
		return ErrCategoryReadFail
	}
	mapping, exist := t.ptcIdToMapping[ptc.ID]
	if exist == false {
		return fmt.Errorf("类型\"%s\"%v", name, ErrCategoryMappingNotExist)
	}
	t.currentTransaction.IncomeExpense = incomeExpense
	t.currentTransaction.CategoryId = mapping.CategoryId
	return nil
}

func (r *AliPayReader) setAmount(t *ReaderTemplate) error {
	amountFloat, err := strconv.ParseFloat(t.currentRow[t.transDataMapping.Amount], 64)
	if err != nil {
		return err
	}
	t.currentTransaction.Amount = int(amountFloat) * 100
	return nil
}

func (r *AliPayReader) setRemark(t *ReaderTemplate) {
	t.currentTransaction.Remark = strings.TrimSpace(t.currentRow[t.transDataMapping.Remark])
}

func (r *AliPayReader) setTradeTime(t *ReaderTemplate) error {
	var err error
	t.currentTransaction.TradeTime, err = time.ParseInLocation(
		t.info.DateFormat,
		t.currentRow[t.transDataMapping.TradeTime], t.location,
	)
	return err
}
