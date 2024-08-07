package lock

import (
	"KeepAccount/initialize"
	"context"
	"errors"
	"time"
)

var (
	currentMode mode

	New             func(key Key) Lock
	NewWithDuration func(key Key, duration time.Duration) Lock

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
		New = func(key Key) Lock {
			return newRedisLock(rdb, string(key), time.Second*10)
		}
		NewWithDuration = func(key Key, duration time.Duration) Lock {
			return newRedisLock(rdb, string(key), duration)
		}
		return
	case redisMode:
		rdb = initialize.LockRdb
		if rdb == nil {
			panic("initialize.LockRdb is nil")
		}
		New = func(key Key) Lock {
			return newMysqlLock(mdb, string(key), time.Second*10)
		}
		NewWithDuration = func(key Key, duration time.Duration) Lock {
			return newMysqlLock(mdb, string(key), duration)
		}
		return
	}
}
