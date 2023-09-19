package initialize

import (
	"KeepAccount/util"
	"fmt"
)

type _redis struct {
	DB       int
	Addr     string
	Password string
}

func (r *_redis) do() error {
	redisCache := &util.RedisCache{DB: r.DB, Addr: r.Addr, Password: r.Password}
	err := redisCache.Init()
	if err == nil {
		Cache = redisCache
		return nil
	}
	//redis 初始化失败
	print(fmt.Sprint("redis 初始化失败 err: %v", err))
	print("初始化 localCache")
	localCache := &util.LocalCache{}
	localCache.Init()
	Cache = localCache
	return nil
}
