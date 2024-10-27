package util

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"KeepAccount/global/constant"
	"github.com/go-redis/redis/v8"
	"github.com/songzhibin97/gkit/cache/local_cache"
)

type Cache interface {
	GetKey(tab constant.CacheTab, unique string) string
	Init() error
	Get(key string) (interface{}, bool)
	GetInt(key string) (int, bool)
	Set(key string, value interface{}, duration time.Duration)
	Increment(key string, number int64) error
	Close() error
	Delete(key string) error
}
type cacheBase struct {
}

func (cb *cacheBase) GetKey(tab constant.CacheTab, unique string) string {
	return string(tab) + "_" + unique
}

type LocalCache struct {
	cacheBase
	local_cache.Cache
}

type RedisCache struct {
	cacheBase
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
			DB:       rc.DB,       // use default db
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
		log.Printf("Error while getting key %s: %v\n", key, err)
		return nil, false
	}
	return val, true
}

func (rc *RedisCache) GetInt(key string) (result int, isSuccess bool) {
	ctx := context.Background()
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return
	} else if err != nil {
		log.Printf("Error while getting key %s: %v\n", key, err)
		return
	}
	result, err = convertToInt(val)
	if err != nil {
		return
	}
	return result, true
}

func (rc *RedisCache) Set(key string, value interface{}, duration time.Duration) {
	ctx := context.Background()
	err := rc.client.Set(ctx, key, value, duration).Err()
	if err != nil {
		log.Printf("Error while setting key %s: %v\n", key, err)
	}
}

func (rc *RedisCache) Increment(key string, number int64) error {
	ctx := context.Background()
	_, err := rc.client.IncrBy(ctx, key, number).Result()
	if err != nil {
		log.Printf("Error while incrementing key %s: %v\n", key, err)
		return err
	}
	return nil
}

func (rc *RedisCache) Delete(key string) error {
	ctx := context.Background()
	_, err := rc.client.Del(ctx, key).Result()
	if err != nil {
		log.Printf("Error while Delete key %s: %v\n", key, err)
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

func (rc *LocalCache) GetInt(key string) (int, bool) {
	val, isSuccess := rc.Get(key)
	if !isSuccess {
		return 0, false
	}
	result, err := convertToInt(val)
	if err != nil {
		return 0, false
	}
	return result, true
}

func (r *LocalCache) Close() error {
	return nil
}

func (r *LocalCache) Delete(key string) error {
	r.Cache.Delete(key)
	return nil
}

func convertToInt(i interface{}) (int, error) {
	switch v := i.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		if v > uint64(int(^uint(0)>>1)) { // 检查是否超出 int 范围
			return 0, fmt.Errorf("uint64 值超出 int 的范围：%d", v)
		}
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		if value, err := strconv.Atoi(v); err == nil {
			return value, nil
		}
		return 0, fmt.Errorf("Unable to convert string to int：%s", v)
	case []byte:
		if value, err := strconv.Atoi(string(v)); err == nil {
			return value, nil
		}
		return 0, fmt.Errorf("Unable to convert []byte to int：%s", v)
	default:
		return 0, fmt.Errorf("Unsupported type：%T", v)
	}
}
