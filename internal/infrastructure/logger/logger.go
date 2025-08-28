// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger/logger.go
package logger

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

type LoggerInterface interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Sync()
	Shutdown(ctx context.Context) error
}

type Logger struct {
	zap *zap.Logger
}

type mode string

const (
	DEV  mode = "development"
	PROD mode = "production"
)

var ErrLoggerInit = errors.New("failed to initialize logger")

func NewLogger(m mode) (LoggerInterface, error) {
	var z *zap.Logger
	var err error

	switch m {
	case DEV:
		z, err = zap.NewDevelopment()
	case PROD:
		z, err = zap.NewProduction()
	default:
		z, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, ErrLoggerInit
	}

	return &Logger{zap: z}, nil
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

func (l *Logger) Sync() {
	l.zap.Sync()
}

func (l *Logger) Shutdown(ctx context.Context) error {
	if err := l.zap.Sync(); err != nil {
		l.zap.Error("logger sync failed", zap.Error(err))
		return err
	}
	return nil
}
