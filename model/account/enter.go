package accountModel

import "KeepAccount/global"

type dao struct {
}

var (
	Dao = &dao{}
)

func init() {
	tables := []interface{}{User{}, UserInvitation{}, Mapping{}}
	for _, table := range tables {
		err := global.GvaDb.AutoMigrate(&table)
		if err != nil {
			panic(err)
		}
	}
}
