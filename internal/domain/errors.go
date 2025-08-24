// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/errors.go
package domain

import "errors"

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidOrder  = errors.New("invalid order")
	ErrDatabaseError = errors.New("database error")
)
