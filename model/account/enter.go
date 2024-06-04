package accountModel

import "KeepAccount/global"

func init() {
	tables := []interface{}{
		&Account{}, &Mapping{},
		&User{}, &UserConfig{}, &UserInvitation{},
	}
	err := global.GvaDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
