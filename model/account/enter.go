package accountModel

import (
	"KeepAccount/global"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"context"
	"fmt"
)

func init() {
	ctx := cus.WithDb(context.TODO(), db.InitDb)
	tables := []interface{}{
		Account{}, Mapping{},
		User{}, UserConfig{}, UserInvitation{},
	}
	err := ctx.GetDb().AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
	err = NewDao(ctx.GetDb()).initRedis()
	if err != nil {
		panic(err)
	}
}

var (
	rdb    = global.GvaRdb
	rdbKey redisKey
)

type redisKey struct {
}

func (r *redisKey) getLocation(id uint) string {
	return fmt.Sprintf("account:%d:location", id)
}
