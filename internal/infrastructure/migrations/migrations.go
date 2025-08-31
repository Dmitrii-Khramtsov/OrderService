// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/migrations/migrations.go
package migrations

import (
	"context"
	"database/sql"
	"fmt"

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
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		l.Error("failed to create migration instance", "error", err)
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	oldVersion, _, _ := m.Version()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		l.Error("failed to apply migrations", "error", err)
		if oldVersion > 0 {
			m.Migrate(oldVersion)
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	newVersion, _, _ := m.Version()

	l.Info("migrations applied successfully",
		"old_version", oldVersion,
		"new_version", newVersion)
	return nil
}
