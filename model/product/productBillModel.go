package productModel

import (
	"time"

	"github.com/ZiRunHua/LeapLedger/global/constant"
	"gorm.io/gorm"
)

type Bill struct {
	ProductKey Key `gorm:"primary_key;"`
	Encoding   constant.Encoding
	StartRow   int
	DateFormat string `gorm:"default:2006-01-02 15:04:05;"`
}

func (b *Bill) TableName() string {
	return "product_bill"
}

type BillImport struct {
	ID          uint
	Status      ImportStatus
	IgnoreCount int
	FinishCount int
	EndTime     time.Time
	gorm.Model
}

func (b *BillImport) TableName() string { return "product_bill_import" }

type ImportStatus int8

const (
	ImportStatusOfReady ImportStatus = iota
	ImportStatusOfImporting
	ImportStatusOfFinish
	ImportStatusOfCancel
)

type BillImportMapping struct {
	ID          uint
	Status      int8
	IgnoreCount int
	FinishCount int
	gorm.Model
}

func (b *BillImportMapping) TableName() string { return "product_bill_import_mapping" }
