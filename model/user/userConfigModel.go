package userModel

import (
	"time"

	"github.com/ZiRunHua/LeapLedger/global"
	"gorm.io/gorm"
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

type Flag uint

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

func (u *TransactionShareConfig) SelectByUserId(userId uint) error {
	u.UserId = userId
	u.DisplayFlags = DISPLAY_FLAGS_DEFAULT
	return global.GvaDb.Where("user_id = ?", userId).FirstOrCreate(&u).Error
}

func (u *TransactionShareConfig) OpenDisplayFlag(flag Flag, db *gorm.DB) error {
	where := db.Where("user_id = ?", u.UserId)
	return where.Model(&u).Update("display_flags", gorm.Expr("display_flags | ?", flag)).Error
}

func (u *TransactionShareConfig) ClosedDisplayFlag(flag Flag, db *gorm.DB) error {
	where := db.Where("user_id = ? AND display_flags & ? > 0", u.UserId, flag)
	return where.Model(&u).Update("display_flags", gorm.Expr("display_flags ^ ?", flag)).Error
}

func (u *TransactionShareConfig) GetFlagStatus(flag Flag) bool {
	return u.DisplayFlags&flag > 0
}

type BillImportConfig struct {
	ConfigBase
}
