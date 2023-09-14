package initialize

import (
	"KeepAccount/util"
)

type _redis struct {
	DB       int    `yaml:"DB"`
	Addr     string `yaml:"Addr"`
	Password string `yaml:"Password"`
}

func (r *_redis) do() error {
	if r.Addr != "" {
		redisCache := &util.RedisCache{DB: r.DB, Addr: r.Addr, Password: r.Password}
		err := redisCache.Init()
		if err == nil {
			Cache = redisCache
			return nil
		}
	}
	localCache := &util.LocalCache{}
	localCache.Init()
	Cache = localCache
	return nil
}
