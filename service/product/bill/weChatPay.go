package bill

import (
	"KeepAccount/global/constant"
	"strconv"
	"strings"
	"time"
)

type WeChatPayReader struct {
	*ReaderTemplate
}

type weChatPayTransactionReader interface {
	TransactionReader
	checkOrderStatus() bool
	setTransCategory() bool
	setAmount() bool
	setRemark()
	setTradeTime() bool
}

func (b *WeChatPayReader) readTransaction() bool {
	if !b.checkOrderStatus() || !b.setTransCategory() || !b.setAmount() || !b.setTradeTime() {
		return false
	}
	b.setRemark()
	return true
}

func (b *WeChatPayReader) checkOrderStatus() bool {
	status := strings.TrimSpace(b.currentRow[b.transDataMapping.TransStatus])
	if status != "支付成功" && status != "已转账" && status != "已收钱" {
		return false
	}
	return true
}

func (b *WeChatPayReader) setTransCategory() bool {
	incomeExpenseStr := strings.TrimSpace(b.currentRow[b.transDataMapping.IncomeExpense])
	var incomeExpense constant.IncomeExpense
	if incomeExpenseStr == "收入" {
		incomeExpense = constant.Income
	} else if incomeExpenseStr == "支出" {
		incomeExpense = constant.Expense
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

func (b *WeChatPayReader) setAmount() bool {
	var amountFloat float64
	amountStr := strings.TrimLeft(b.currentRow[b.transDataMapping.Amount], "¥")
	amountFloat, b.err = strconv.ParseFloat(amountStr, 64)
	if b.err != nil {
		return false
	} else {
		b.currentTransaction.Amount = int(amountFloat) * 100
	}
	return true
}

func (b *WeChatPayReader) setRemark() {
	b.currentTransaction.Remark = strings.TrimSpace(b.currentRow[b.transDataMapping.Remark])
}

func (b *WeChatPayReader) setTradeTime() bool {
	date := b.currentRow[b.transDataMapping.TradeTime]
	if b.currentTransaction.TradeTime, b.err = time.Parse(b.info.DateFormat, date); b.err != nil {
		return false
	}
	return true
}
