// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/database.go
package factory

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	infrarepo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
)

func NewDatabase(cfg *config.Config, l domainrepo.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
	if err != nil {
		l.Error("failed to connect to db", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", domainrepo.ErrDatabaseConnectionFailed, err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db, nil
}

func NewOrderRepository(cfg *config.Config, db *sqlx.DB, l domainrepo.Logger) (domainrepo.OrderRepository, error) {
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
