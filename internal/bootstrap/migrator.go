// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/migrator.go
package bootstrap

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/migrations"
)

func RunMigrations(ctx context.Context, db *sqlx.DB, migrationsPath string, l logger.LoggerInterface, cfg *config.Config) error {
	operation := func() error {
		return migrations.RunMigrations(ctx, db.DB, migrationsPath, l)
	}

	retryPolicy := backoff.NewExponentialBackOff()
	retryPolicy.MaxElapsedTime = cfg.Retry.MaxElapsedTime

	if err := backoff.Retry(operation, retryPolicy); err != nil {
		l.Error("failed to run migrations after retries", zap.Error(err))
		return err
	}

	l.Info("migrations completed successfully")
	return nil
}
