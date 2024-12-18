package userModel

import "github.com/ZiRunHua/LeapLedger/global/db"

func init() {
	tables := []interface{}{
		User{}, UserClientWeb{}, UserClientAndroid{}, UserClientIos{}, Tour{},
		Friend{}, FriendInvitation{},
		TransactionShareConfig{},
		Log{},
	}
	err := db.InitDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
