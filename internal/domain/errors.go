// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/errors.go
package domain

import "errors"

var (
	ErrInvalidOrder  = errors.New("invalid order")
	ErrOrderNotFound = errors.New("order not found")
)
