package userModel

import "github.com/ZiRunHua/LeapLedger/global/db"

func init() {
	tables := []interface{}{
		User{}, UserClientWeb{}, UserClientAndroid{}, UserClientIos{}, Tour{},
		Friend{}, FriendInvitation{},
		TransactionShareConfig{}, BillImportConfig{},
		Log{},
	}
	err := db.InitDb.AutoMigrate(tables...)
	if err != nil {
		panic(err.Error())
	}
	for config := range DefaultConfigs.Iterator(0) {
		err = db.InitDb.AutoMigrate(config)
		if err != nil {
			panic(err)
		}
	}
}
