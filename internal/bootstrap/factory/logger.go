// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/logger.go
package factory

import (
	"os"

	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func NewLogger(cfg *config.Config) (domainrepo.Logger, error) {
	mode := logger.DEV
	if envLogMode := os.Getenv("LOG_MODE"); envLogMode == "production" {
		mode = logger.PROD
	}
	return logger.NewLogger(mode)
}
