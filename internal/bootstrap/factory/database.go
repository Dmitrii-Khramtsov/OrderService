// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/database.go
package factory

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	infrarepo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
)

func NewDatabase(cfg *config.Config, l domainrepo.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
	if err != nil {
		l.Error("failed to connect to db", "error", err)
		return nil, fmt.Errorf("%w: %v", infrarepo.ErrDatabaseConnectionFailed, err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	_, err = db.Exec(fmt.Sprintf(
		"SET statement_timeout = %d; SET idle_in_transaction_session_timeout = %d;",
		cfg.Database.StatementTimeout.Milliseconds(),
		cfg.Database.IdleInTxSessionTimeout.Milliseconds(),
	))
	if err != nil {
		l.Error("failed to set timeouts", "error", err)
	}

	return db, nil
}

func NewOrderRepository(cfg *config.Config, db *sqlx.DB, l domainrepo.Logger) (domainrepo.OrderRepository, error) {
	baseRepo, err := infrarepo.NewPostgresOrderRepository(db, l, cfg.Database.StatementTimeout)
	if err != nil {
		return nil, err
	}

	retryConfig := &infrarepo.RetryConfig{
		MaxElapsedTime:      cfg.Kafka.Retry.MaxElapsedTime,
		InitialInterval:     cfg.Kafka.Retry.InitialInterval,
		RandomizationFactor: cfg.Kafka.Retry.RandomizationFactor,
		Multiplier:          cfg.Kafka.Retry.Multiplier,
		MaxInterval:         cfg.Kafka.Retry.MaxInterval,
	}

	return infrarepo.NewRetryingOrderRepository(baseRepo, l, retryConfig), nil
}
