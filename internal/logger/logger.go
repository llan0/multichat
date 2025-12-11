package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}

type zapLogger struct {
	*zap.Logger
}

func (z *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{z.Logger.With(fields...)}
}

func (z *zapLogger) Sync() error {
	return z.Logger.Sync()
}

func New() (Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &zapLogger{l}, nil
}

// no-op logger for testing
func NewNop() Logger {
	return &zapLogger{zap.NewNop()}
}
