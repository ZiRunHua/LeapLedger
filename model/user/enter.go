package userModel

import "KeepAccount/global"

type dao struct {
}

func init() {
	tables := []interface{}{TransactionShareConfig{}, Friend{}, FriendInvitation{}}
	for _, table := range tables {
		err := global.GvaDb.AutoMigrate(&table)
		if err != nil {
			panic(err)
		}
	}
}
