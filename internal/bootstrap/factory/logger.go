package factory

import (
	"os"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func NewLogger(cfg *config.Config) (logger.LoggerInterface, error) {
	mode := logger.DEV
	if envLogMode := os.Getenv("LOG_MODE"); envLogMode == "production" {
		mode = logger.PROD
	}
	return logger.NewLogger(mode)
}
