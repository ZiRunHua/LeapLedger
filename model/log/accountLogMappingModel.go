package logModel

import (
	"gorm.io/gorm"
)

type AccountLogMapping struct {
	ID        uint `gorm:"primarykey"`
	AccountId uint `gorm:"index:idx_account_id"`
	LogTable  string
	LogId     uint
	gorm.Model
}

func (a *AccountLogMapping) TableName() string {
	return "account_log_mapping"
}
