// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/migrations/migrations.go
package migrations

import (
	"context"
	"database/sql"

	repo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"go.uber.org/zap"
)

func RunMigrations(ctx context.Context, db *sql.DB, l logger.LoggerInterface) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS orders (
			order_uid TEXT PRIMARY KEY,
			track_number TEXT NOT NULL,
			entry TEXT,
			locale TEXT,
			internal_signature TEXT,
			customer_id TEXT,
			delivery_service TEXT,
			shardkey TEXT,
			sm_id INT,
			date_created TIMESTAMP,
			oof_shard TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS delivery (
			order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
			name TEXT, phone TEXT, zip TEXT, city TEXT, address TEXT, region TEXT, email TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS payment (
			order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
			transaction TEXT, request_id TEXT, currency TEXT, provider TEXT,
			amount INT, payment_dt BIGINT, bank TEXT, delivery_cost INT, goods_total INT, custom_fee INT
		);`,
		`CREATE TABLE IF NOT EXISTS items (
			chrt_id BIGINT,
			order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
			track_number TEXT, price INT, rid TEXT, name TEXT, sale INT, size TEXT,
			total_price INT, nm_id BIGINT, brand TEXT, status INT,
			PRIMARY KEY (chrt_id, order_uid)
		);`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			l.Error("migration statement failed", zap.Error(err))
			return repo.ErrMigrationFailed
		}
	}

	l.Info("all tables created or already exist")

	rows, err := db.QueryContext(ctx, `SELECT table_name FROM information_schema.tables WHERE table_schema='public';`)
	if err != nil {
		l.Error("failed to list tables", zap.Error(err))
		return repo.ErrMigrationFailed
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			l.Error("failed to scan table name", zap.Error(err))
			return repo.ErrMigrationFailed
		}
		l.Info("table found", zap.String("table", table))
	}

	return nil
}
