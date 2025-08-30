// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/migrations/migrations.go
package migrations

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

// контекст не участвует, оставлен для консистентности
func RunMigrations(ctx context.Context, db *sql.DB, migrationsPath string, l domainrepo.Logger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		l.Error("failed to create migration driver", "error", err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		l.Error("failed to create migration instance", "error", err)
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		l.Error("failed to apply migrations", "error", err)
		return err
	}

	l.Info("migrations applied successfully")
	return nil
}
