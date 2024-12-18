package initialize

import (
	"os"
	"path/filepath"

	"github.com/ZiRunHua/LeapLedger/global/constant"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type _logger struct {
	encoder zapcore.Encoder
}

var (
	_requestLogPath = filepath.Join(constant.RootDir, "log", "request.log")
	_errorLogPath   = filepath.Join(constant.RootDir, "log", "error.log")
	_panicLogPath   = filepath.Join(constant.RootDir, "log", "panic.log")
)

func (l *_logger) do() error {
	l.setEncoder()
	var err error
	if RequestLogger, err = l.New(_requestLogPath); err != nil {
		return err
	}
	if ErrorLogger, err = l.New(_errorLogPath); err != nil {
		return err
	}
	if PanicLogger, err = l.New(_panicLogPath); err != nil {
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

func (l *_logger) New(path string) (*zap.Logger, error) {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	writeSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(l.encoder, writeSyncer, zapcore.DebugLevel)
	return zap.New(core), nil
}
