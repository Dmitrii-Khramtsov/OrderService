// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger/logger.go
package logger

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

type Logger struct {
	zap *zap.Logger
}

type mode string

const (
	DEV  mode = "development"
	PROD mode = "production"
)

var ErrLoggerInit = errors.New("failed to initialize logger")

func NewLogger(m mode) (
	domainrepo.Logger, error) {
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

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.zap.Debug(msg, l.convertFields(fields)...)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.zap.Info(msg, l.convertFields(fields)...)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.zap.Warn(msg, l.convertFields(fields)...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.zap.Error(msg, l.convertFields(fields)...)
}

func (l *Logger) Sync() error {
	return l.zap.Sync()
}

func (l *Logger) Shutdown(ctx context.Context) error {
	if err := l.zap.Sync(); err != nil {
		l.zap.Error("logger sync failed", zap.Error(err))
		return err
	}
	return nil
}

func (l *Logger) convertFields(fields []interface{}) []zap.Field {
	var zapFields []zap.Field

	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			l.zap.Error("odd number of arguments passed to logger")
			break
		}

		key, ok := fields[i].(string)
		if !ok {
			l.zap.Error("logger key is not a string")
			continue
		}

		value := fields[i+1]

		switch v := value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(key, v))
		case int:
			zapFields = append(zapFields, zap.Int(key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		case bool:
			zapFields = append(zapFields, zap.Bool(key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(key, v))
		default:
			zapFields = append(zapFields, zap.String(key, fmt.Sprintf("%v", v)))
		}
	}

	return zapFields
}
