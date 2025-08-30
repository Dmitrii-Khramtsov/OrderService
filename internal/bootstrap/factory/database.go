// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/database.go
package factory

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	repo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	infrarepo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func NewDatabase(cfg *config.Config, l logger.LoggerInterface) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
	if err != nil {
		l.Error("failed to connect to db", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", repo.ErrDatabaseConnectionFailed, err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db, nil
}

func NewOrderRepository(cfg *config.Config, db *sqlx.DB, l logger.LoggerInterface) (repo.OrderRepository, error) {
	baseRepo, err := infrarepo.NewPostgresOrderRepository(db, l)
	if err != nil {
		return nil, err
	}

	retryConfig := &infrarepo.RetryConfig{
		MaxElapsedTime:      cfg.Retry.MaxElapsedTime,
		InitialInterval:     cfg.Retry.InitialInterval,
		RandomizationFactor: cfg.Retry.RandomizationFactor,
		Multiplier:          cfg.Retry.Multiplier,
		MaxInterval:         cfg.Retry.MaxInterval,
	}

	return infrarepo.NewRetryingOrderRepository(baseRepo, l, retryConfig), nil
}
