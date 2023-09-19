package util

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"time"
)

type Cache interface {
	Init() error
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, duration time.Duration)
	Increment(key string, number int64) error
	Close() error
}

type LocalCache struct {
	local_cache.Cache
}

type RedisCache struct {
	client   *redis.Client
	DB       int
	Addr     string
	Password string
}

func (rc *RedisCache) Init() error {
	client := redis.NewClient(
		&redis.Options{
			Addr:     rc.Addr,
			Password: rc.Password, // no password set
			DB:       rc.DB,       // use default DB
		},
	)
	_, err := client.Ping(context.Background()).Result()
	rc.client = client
	return err
}

func (rc *RedisCache) Get(key string) (interface{}, bool) {
	ctx := context.Background()
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		fmt.Printf("Error while getting key %s: %v\n", key, err)
		return nil, false
	}
	return val, true
}

func (rc *RedisCache) Set(key string, value interface{}, duration time.Duration) {
	ctx := context.Background()
	err := rc.client.Set(ctx, key, value, duration).Err()
	if err != nil {
		fmt.Printf("Error while setting key %s: %v\n", key, err)
	}
}

func (rc *RedisCache) Increment(key string, number int64) error {
	ctx := context.Background()
	_, err := rc.client.IncrBy(ctx, key, number).Result()
	if err != nil {
		fmt.Printf("Error while incrementing key %s: %v\n", key, err)
		return err
	}
	return nil
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

func (r *LocalCache) Init() error {
	r.Cache = local_cache.NewCache(
		local_cache.SetDefaultExpire(time.Hour * 2),
	)
	return nil
}

func (r *LocalCache) Close() error {
	return nil
}
