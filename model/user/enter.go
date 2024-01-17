package userModel

import "KeepAccount/global"

type dao struct {
}

var (
	Dao = &dao{}
)

func init() {
	err := global.GvaDb.AutoMigrate(&TransactionShareConfig{})
	if err != nil {
		panic(err)
	}
}
