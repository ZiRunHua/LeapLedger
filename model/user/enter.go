package userModel

import (
	"KeepAccount/util/gormFunc"
	"gorm.io/gorm"
)

func CurrentInit(db *gorm.DB) error {
	tables := []interface{}{
		User{}, UserClientWeb{}, UserClientAndroid{}, UserClientIos{}, Tour{},
		Friend{}, FriendInvitation{},
		TransactionShareConfig{},
		Log{},
	}
	err := db.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		_ = gormFunc.AlterIdToHeader(table, db)
	}
	return nil
}
