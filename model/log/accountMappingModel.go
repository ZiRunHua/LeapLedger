package logModel

import (
	"gorm.io/gorm"
)

type AccountMappingLogData struct {
	MainId    uint
	RelatedId uint
}

func (m *AccountMappingLogData) Record(baseLog BaseAccountLog, tx *gorm.DB) (AccountLogger, error) {
	log := AccountMappingLog{BaseAccountLog: baseLog, Data: *m}
	return &log, tx.Create(&log).Error
}

type AccountMappingLog struct {
	BaseAccountLog `gorm:"embedded"`
	Data           AccountMappingLogData `gorm:"embedded;embeddedPrefix:data_"`
}

func (a *AccountMappingLog) TableName() string {
	return "log_account_mapping"
}

func (a *AccountMappingLog) RecordMapping(tx *gorm.DB) (AccountLogMapping, error) {
	return NewDao(tx).RecordAccountLogMapping(a)
}
