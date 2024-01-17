package userModel

import (
	"KeepAccount/global"
	"gorm.io/gorm"
)

type TransactionShareConfig struct {
	gorm.Model
	UserId       uint `gorm:"type:int;unsigned;comment:'用户ID';unique"`
	DisplayFlags Flag `gorm:"type:smallint;unsigned;comment:'展示字段标志'"`
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
	where := db.Where("user_id = ? AND display_flags & ? >0", u.UserId, flag)
	return where.Model(&u).Update("display_flags", gorm.Expr("display_flags ^ ?", flag)).Error
}

func (u *TransactionShareConfig) GetFlagStatus(flag Flag) bool {
	return u.DisplayFlags&flag > 0
}
