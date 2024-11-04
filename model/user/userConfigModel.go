package userModel

import (
	"time"
)

func newDefaultConfig(userId uint) []Config {
	base := ConfigBase{UserId: userId}
	return []Config{
		&TransactionShareConfig{ConfigBase: base, DisplayFlags: DISPLAY_FLAGS_DEFAULT},
		&BillImportConfig{ConfigBase: base},
	}
}

type Config interface {
	GetUserId() uint
	SetUserId(uint)
}

type ConfigBase struct {
	Config
	UserId    uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (cb *ConfigBase) GetUserId() uint   { return cb.UserId }
func (cb *ConfigBase) SetUserId(id uint) { cb.UserId = id }

type TransactionShareConfig struct {
	ConfigBase
	DisplayFlags Flag `gorm:"comment:'展示字段标志'"`
}

type Flag = uint

const (
	FLAG_AMOUNT Flag = 1 << iota
	FLAG_CATEGORY
	FLAG_TRADE_TIME
	FLAG_ACCOUNT
	FLAG_CREATE_TIME
	FLAG_UPDATE_TIME
	FLAG_REMARK
)

const DISPLAY_FLAGS_DEFAULT = FLAG_AMOUNT + FLAG_CATEGORY + FLAG_TRADE_TIME + FLAG_ACCOUNT + FLAG_REMARK

func (u *TransactionShareConfig) TableName() string {
	return "user_transaction_share_config"
}

func (u *TransactionShareConfig) GetFlagStatus(flag Flag) bool {
	return u.DisplayFlags&flag > 0
}

type BillImportConfig struct {
	ConfigBase
}
