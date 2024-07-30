package lock

import (
	"KeepAccount/initialize"
	"context"
	"errors"
	"time"
)

var (
	currentMode mode

	NewLock             func(key string) Lock
	NewLockWithDuration func(key string, duration time.Duration) Lock

	ErrLockNotExist = errors.New("lock not exist")
	ErrLockOccupied = errors.New("lock occupied")
)

type mode string

const (
	mysqlMode mode = "mysql"
	redisMode mode = "redis"
)

type Lock interface {
	Lock(context.Context) error
	Release(context.Context) error
}

func init() {
	currentMode = mode(initialize.Config.System.LockMode)
	updatePublicFunc()
}
func updatePublicFunc() {
	switch currentMode {
	case mysqlMode:
		mdb = initialize.Db
		err := mdb.AutoMigrate(&lockTable{})
		if err != nil {
			panic(err)
		}
		NewLock = func(key string) Lock {
			return newRedisLock(rdb, key, time.Second*10)
		}
		NewLockWithDuration = func(key string, duration time.Duration) Lock {
			return newRedisLock(rdb, key, duration)
		}
		return
	case redisMode:
		rdb = initialize.LockRdb
		if rdb == nil {
			panic("initialize.LockRdb is nil")
		}
		NewLock = func(key string) Lock {
			return newMysqlLock(mdb, key, time.Second*10)
		}
		NewLockWithDuration = func(key string, duration time.Duration) Lock {
			return newMysqlLock(mdb, key, duration)
		}
		return
	}
}
