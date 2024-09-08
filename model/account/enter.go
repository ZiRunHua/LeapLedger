package accountModel

import (
	"KeepAccount/global/db"
)

func init() {
	tables := []interface{}{
		Account{}, Mapping{},
		User{}, UserConfig{}, UserInvitation{},
	}
	err := db.InitDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
