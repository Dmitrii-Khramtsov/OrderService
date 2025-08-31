// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository/errors.go
package repository

import "errors"

var (
	ErrDatabaseConnectionFailed = errors.New("failed to connect to database")
	ErrMigrationFailed          = errors.New("migration failed")
	ErrQueryFailed              = errors.New("database query failed")
	ErrNotFound                 = errors.New("record not found")
	ErrInsertFailed             = errors.New("failed to insert record")
	ErrUpdateFailed             = errors.New("failed to update record")
	ErrDeleteFailed             = errors.New("failed to delete record")
	ErrTransactionFailed        = errors.New("transaction failed")
	ErrOrderSaveFailed          = errors.New("failed to save order")
	ErrOrderDeleteFailed        = errors.New("failed to delete order")
	ErrOrderClearFailed         = errors.New("failed to clear orders")
)
