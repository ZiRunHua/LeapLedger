package redisTool

import (
	"KeepAccount/global"
	"context"
	"strconv"
)

var (
	rdb = global.GvaRdb
)

func GetInt(key string, ctx context.Context) (value int, err error) {
	str, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return
	}
	return strconv.Atoi(str)
}
