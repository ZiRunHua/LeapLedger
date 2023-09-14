package initialize

import (
	"KeepAccount/global"
	"context"
	"github.com/go-redis/redis/v8"
)

func Redis() {
	redisCfg := global.GvaConfig.Redis
	client := redis.NewClient(
		&redis.Options{
			Addr:     redisCfg.Addr,
			Password: redisCfg.Password, // no password set
			DB:       redisCfg.DB,       // use default DB
		},
	)
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		print(err)
	} else {
		print(pong)
		global.GvaRedis = client
	}
}
