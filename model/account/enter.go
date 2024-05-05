package accountModel

import "KeepAccount/global"

func init() {
	tables := []interface{}{User{}, UserInvitation{}, Mapping{}, UserConfig{}}
	for _, table := range tables {
		err := global.GvaDb.AutoMigrate(&table)
		if err != nil {
			panic(err)
		}
	}
}
