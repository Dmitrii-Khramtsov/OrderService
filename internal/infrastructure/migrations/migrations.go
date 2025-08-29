// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/migrations/migrations.go
package migrations

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// контекст не участвует, оставлен для консистентности
func RunMigrations(ctx context.Context, db *sql.DB, migrationsPath string, l logger.LoggerInterface) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		l.Error("failed to create migration driver", zap.Error(err))
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		l.Error("failed to create migration instance", zap.Error(err))
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		l.Error("failed to apply migrations", zap.Error(err))
		return err
	}

	l.Info("migrations applied successfully")
	return nil
}
