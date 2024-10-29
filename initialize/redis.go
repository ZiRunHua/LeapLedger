package initialize

import (
	"context"
	"log"
	"time"

	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/util"
	"github.com/go-redis/redis/v8"
)

type _redis struct {
	Addr     string `yaml:"Addr"`
	Password string `yaml:"Password"`
	Db       int    `yaml:"Db"`
	LockDb   int    `yaml:"LockDb"`
}

var Rdb, LockRdb *redis.Client

func (r *_redis) do() error {
	if len(r.Addr) == 0 {
		return nil
	}
	var err error
	Rdb, err = r.getNewRedisClient("", r.Db)
	if err != nil {
		return err
	}
	LockRdb, err = r.getNewRedisClient("lock", r.LockDb)
	if err != nil {
		return err
	}
	Cache = &util.RedisCache{DB: r.Db, Addr: r.Addr, Password: r.Password}
	return Cache.Init()
}

func (r *_redis) getNewRedisClient(name string, dbNum int) (*redis.Client, error) {
	connect := func() (*redis.Client, error) {
		db := redis.NewClient(&redis.Options{Addr: r.Addr, Password: r.Password, DB: dbNum})
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
		defer cancel()
		return db, db.Ping(ctx).Err()
	}
	db, err := reconnection[*redis.Client](connect, 3)
	if err != nil {
		return db, err
	}
	if Config.Mode == constant.Debug {
		db.AddHook(&RedisHook{name: name})
	}
	return db, err
}

type RedisHook struct {
	name string
}

func (rh RedisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if len(rh.name) == 0 {
		log.Printf("exec  => <%s>\n", cmd)
	} else {
		log.Printf("%s exec  => <%s>\n", rh.name, cmd)
	}
	return ctx, nil
}

func (rh RedisHook) AfterProcess(_ context.Context, cmd redis.Cmder) error {
	if len(rh.name) == 0 {
		log.Printf("finish => <%s>\n", cmd)
	} else {
		log.Printf("%s finish => <%s>\n", rh.name, cmd)
	}
	return nil
}

func (rh RedisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	if len(rh.name) == 0 {
		log.Printf("pipeline exec   => %v\n", cmds)
	} else {
		log.Printf("%s pipeline exec   => %v\n", rh.name, cmds)
	}
	return ctx, nil
}

func (rh RedisHook) AfterProcessPipeline(_ context.Context, cmds []redis.Cmder) error {
	if len(rh.name) == 0 {
		log.Printf("pipeline finish => %v\n", cmds)
	} else {
		log.Printf("%s pipeline finish => %v\n", rh.name, cmds)
	}
	return nil
}
