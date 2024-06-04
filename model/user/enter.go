package userModel

import "KeepAccount/global"

func init() {
	tables := []interface{}{
		&User{}, &UserClientWeb{}, &UserClientAndroid{}, &UserClientIos{},
		&Friend{}, &FriendInvitation{},
		&TransactionShareConfig{},
		&Log{},
	}
	err := global.GvaDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
