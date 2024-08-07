package productModel

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	commonModel "KeepAccount/model/common"
	"database/sql"
	"time"
)

type TransactionCategoryMapping struct {
	AccountId  uint `gorm:"uniqueIndex:account_ptc_mapping,priority:1"`
	CategoryId uint `gorm:"uniqueIndex:category_ptc_mapping,priority:1"`
	PtcId      uint `gorm:"uniqueIndex:account_ptc_mapping,priority:2;uniqueIndex:category_ptc_mapping,priority:2"`
	ProductKey string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	commonModel.BaseModel
}

func (tcm *TransactionCategoryMapping) TableName() string {
	return "product_transaction_category_mapping"
}

func (tcm *TransactionCategoryMapping) IsEmpty() bool {
	return tcm == nil || tcm.AccountId == 0
}

func (tcm *TransactionCategoryMapping) GetPtcIdMapping(
	account *accountModel.Account, productKey KeyValue,
) (result map[uint]TransactionCategoryMapping, err error) {
	db := global.GvaDb
	rows, err := db.Model(&TransactionCategoryMapping{}).Where(
		"account_id = ? AND product_key = ?", account.ID, productKey,
	).Rows()
	defer func(rows *sql.Rows) {
		if err != nil {
			_ = rows.Close()
		} else {
			err = rows.Close()
		}
	}(rows)
	if err != nil {
		return
	}
	row, result := TransactionCategoryMapping{}, map[uint]TransactionCategoryMapping{}
	for rows.Next() {
		err = db.ScanRows(rows, &row)
		if err != nil {
			return
		}
		result[row.PtcId] = row
	}
	return
}
