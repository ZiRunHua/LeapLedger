package initialize

import (
	"KeepAccount/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	requestLogPath = "log/request.log"
	errorLogPath   = "log/error.log"
	panicLogPath   = "log/panic.log"
)

func Logger() {
	encoder := getEncoder()
	initRequestLogger(encoder)
	initErrorLogger(encoder)
	initPanicLogger(encoder)
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func initRequestLogger(encoder zapcore.Encoder) {
	file, err := os.OpenFile(requestLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	writeSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	global.RequestLogger = zap.New(core, zap.AddCaller())
}

func initErrorLogger(encoder zapcore.Encoder) {
	file, err := os.OpenFile(errorLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	writeSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	global.ErrorLogger = zap.New(core)
}

func initPanicLogger(encoder zapcore.Encoder) {
	file, err := os.OpenFile(panicLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	writeSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	global.PanicLogger = zap.New(core, zap.AddCaller())
}
