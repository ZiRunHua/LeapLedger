package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

func GetNewZapLogger(logPath string) (*zap.Logger, error) {
	logDir := filepath.Dir(logPath)
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(logFile),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)
	logger := zap.New(core, zap.AddCaller())
	return logger, nil
}
