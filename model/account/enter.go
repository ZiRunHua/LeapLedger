package accountModel

import (
	"context"
	"fmt"
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/cus"
	"github.com/ZiRunHua/LeapLedger/global/db"
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
