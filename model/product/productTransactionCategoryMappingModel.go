package productModel

import (
	accountModel "KeepAccount/model/account"
	commonModel "KeepAccount/model/common"
	"time"
)

type TransactionCategoryMapping struct {
	AccountID  uint
	CategoryID uint
	PtcID      uint
	ProductKey string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	commonModel.BaseModel
}

func (tcm *TransactionCategoryMapping) TableName() string {
	return "product_transaction_category_mapping"
}

func (tcm *TransactionCategoryMapping) IsEmpty() bool {
	return tcm == nil || tcm.AccountID == 0
}

func (tcm *TransactionCategoryMapping) GetPtcIdMapping(
	account *accountModel.Account, productKey KeyValue,
) (result map[uint]TransactionCategoryMapping, err error) {
	db := tcm.GetDb()
	rows, err := db.Model(&TransactionCategoryMapping{}).Where(
		"account_id = ? AND product_key = ?", account.ID, productKey,
	).Rows()
	defer rows.Close()

	row, result := TransactionCategoryMapping{}, map[uint]TransactionCategoryMapping{}
	for rows.Next() {
		db.ScanRows(rows, &row)
		result[row.PtcID] = row
	}
	return
}
