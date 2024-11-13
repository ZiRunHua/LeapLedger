package transactionModel

import (
	"time"

	"gorm.io/gorm"
)

type Import struct {
	ID          uint
	Status      ImportStatus
	IgnoreCount int
	FinishCount int
	EndTime     time.Time
	gorm.Model
}

func (b *Import) TableName() string { return "transaction_import" }

type ImportStatus int8

const (
	ImportStatusOfReady ImportStatus = iota
	ImportStatusOfImporting
	ImportStatusOfFinish
	ImportStatusOfCancel
)

type ImportDetail struct {
	ID       uint
	ImportId uint
	TransId  uint
	gorm.Model
}

func (bt *ImportDetail) TableName() string { return "transaction_import_detail" }
