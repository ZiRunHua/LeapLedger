package lock

import (
	"KeepAccount/initialize"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"time"
)

var (
	rdb = initialize.LockRdb
)

type RedisLock struct {
	client     *redis.Client
	key        string
	value      string
	expiration time.Duration
}

func newRedisLock(client *redis.Client, key string, expiration time.Duration) *RedisLock {
	return &RedisLock{
		client:     client,
		key:        key,
		value:      uuid.New().String(),
		expiration: expiration,
	}
}

func (rl *RedisLock) Lock(ctx context.Context) error {
	success, err := rl.client.SetNX(ctx, rl.key, rl.value, rl.expiration).Result()
	if err != nil {
		return err
	}
	if false == success {
		return ErrLockOccupied
	}
	return nil
}

func (rl *RedisLock) Release(ctx context.Context) error {
	script := `
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    else
        return 0
    end
    `
	result, err := rl.client.Eval(ctx, script, []string{rl.key}, rl.value).Int()
	if err != nil {
		return err
	} else if result == 0 {
		return ErrLockNotExist
	}
	return nil
}
