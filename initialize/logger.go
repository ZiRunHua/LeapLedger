package initialize

import (
	"os"

	"KeepAccount/global/constant"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type _logger struct {
	encoder zapcore.Encoder
}

const (
	_requestLogPath = constant.WORK_PATH + "/log/request.log"
	_errorLogPath   = constant.WORK_PATH + "/log/error.log"
	_panicLogPath   = constant.WORK_PATH + "/log/panic.log"
)

func (l *_logger) do() error {
	l.setEncoder()
	var err error
	if RequestLogger, err = l.initLogger(_requestLogPath); err != nil {
		return err
	}
	if ErrorLogger, err = l.initLogger(_errorLogPath); err != nil {
		return err
	}
	if PanicLogger, err = l.initLogger(_panicLogPath); err != nil {
		return err
	}
	return nil
}

func (l *_logger) setEncoder() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	l.encoder = zapcore.NewConsoleEncoder(encoderConfig)
}

func (l *_logger) initLogger(path string) (*zap.Logger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	writeSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(l.encoder, writeSyncer, zapcore.DebugLevel)
	return zap.New(core), nil
}
