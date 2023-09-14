package productModel

import (
	"KeepAccount/global"
	commonModel "KeepAccount/model/common"
)

type BillHeader struct {
	ID         uint
	ProductKey string
	Name       string
	Type       BillHeaderType
	commonModel.BaseModel
}
type BillHeaderType string

const (
	TransTime     BillHeaderType = "trans_time"
	TransCategory BillHeaderType = "trans_category"
	Remark        BillHeaderType = "remark"
	IncomeExpense BillHeaderType = "income_expense"
	Amount        BillHeaderType = "amount"
	OrderNumber   BillHeaderType = "order_number"
	TransStatus   BillHeaderType = "trans_status"
)

func (b *BillHeader) TableName() string {
	return "product_bill_header"
}

func (b *BillHeader) IsEmpty() bool {
	return b.ID == 0
}

func (tc *BillHeader) GetNameMap(productKey string) (
	NameMap map[string]BillHeader, err error,
) {
	var billHeader BillHeader
	rows, err := global.GvaDb.Model(&billHeader).Where(
		"product_key = ? ", productKey,
	).Rows()
	defer rows.Close()
	if err != nil {
		return
	}
	NameMap = map[string]BillHeader{}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &billHeader)
		if err != nil {
			return
		}
		NameMap[billHeader.Name] = billHeader
	}
	return
}
