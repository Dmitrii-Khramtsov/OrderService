// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger/logger.go
package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Logger struct {
	writer   io.Writer
	minLevel LogLevel
}

func NewLogger(writer io.Writer, minLevel LogLevel) *Logger {
	return &Logger{
		writer: writer,
		minLevel: minLevel,
	}
}

func NewDefaultLogger(minLevel LogLevel) *Logger {
	return NewLogger(os.Stdout, minLevel)
}

type LogLevel int

const (
	INFO LogLevel = iota
	WARN
	ERROR
)

func (lev LogLevel) String() string {
	switch lev {
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) Info(msg string) {
	l.Log(INFO, msg)
}

func (l *Logger) Warn(msg string) {
    l.Log(WARN, msg)
}

func (l *Logger) Error(msg string) {
    l.Log(ERROR, msg)
}

func (l *Logger) Infof(format string, args... interface{}) {
	l.Log(INFO, fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(format string, args... interface{}) {
	l.Log(WARN, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args... interface{}) {
	l.Log(ERROR, fmt.Sprintf(format, args...))
}

func (l *Logger) Log(level LogLevel, msg string) {
	if level < l.minLevel {
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	fmt.Fprintf(l.writer, "[%s] [%s] %s\n", now, level.String(), msg)
}
