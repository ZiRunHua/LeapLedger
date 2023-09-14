package global

import (
	"KeepAccount/initialize"
	"go.uber.org/zap"
)

var (
	GvaDb  = initialize.Db
	Config = initialize.Config
	Cache  = initialize.Cache
)

var (
	RequestLogger *zap.Logger
	ErrorLogger   *zap.Logger
	PanicLogger   *zap.Logger
)

func init() {
	GvaDb = initialize.Db
	Config = initialize.Config
	Cache = initialize.Cache

	RequestLogger = initialize.RequestLogger
	ErrorLogger = initialize.ErrorLogger
	PanicLogger = initialize.PanicLogger
}
