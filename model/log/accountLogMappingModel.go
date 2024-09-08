package logModel

import (
	"gorm.io/gorm"
	"time"
)

type AccountLogMapping struct {
	ID        uint `gorm:"primarykey"`
	AccountId uint `gorm:"index:idx_account_id"`
	LogTable  string
	LogId     uint
	CreatedAt time.Time      `gorm:"type:TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:TIMESTAMP"`
}

func (a *AccountLogMapping) TableName() string {
	return "account_log_mapping"
}
