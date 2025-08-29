package factory

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	repo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func NewDatabase(cfg *config.Config, l logger.LoggerInterface) *sqlx.DB {
	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
	if err != nil {
		l.Error("failed to connect to db", zap.Error(err))
		panic(repo.ErrDatabaseConnectionFailed) // можно вернуть ошибку, но в DI часто паникуют
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db
}

func NewOrderRepository(db *sqlx.DB, l logger.LoggerInterface) (repo.OrderRepository, error) {
	return repo.NewPostgresOrderRepository(db, l)
}
