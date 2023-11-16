package initialize

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type _logger struct {
	Path    string          `yaml:"Path"`   // 日志路径
	Level   string          `yaml:"Level"`  // 日志级别
	Format  string          `yaml:"Format"` // 日志格式
	encoder zapcore.Encoder // 这是一个字段，不需要YAML标签
}

const (
	_requestLogPath = "log/request.log"
	_errorLogPath   = "log/error.log"
	_panicLogPath   = "log/panic.log"
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
