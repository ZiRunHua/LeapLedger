package global

import (
	"KeepAccount/config"
	"github.com/go-redis/redis/v8"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	GvaDb      *gorm.DB
	GvaRedis   *redis.Client
	GvaConfig  config.Server
	BlackCache local_cache.Cache
)

var (
	RequestLogger *zap.Logger
	ErrorLogger   *zap.Logger
	PanicLogger   *zap.Logger
)

// IncomeExpense 收支类型
type IncomeExpense string

const (
	Income  IncomeExpense = "income"
	Expense IncomeExpense = "expense"
)

// Client 客户端
type Client string

const (
	Web     Client = "web"
	Android Client = "android"
	Ios     Client = "ios"
)

type Encoding string

const (
	GBK  Encoding = "GBK"
	UTF8 Encoding = "UTF8"
)
