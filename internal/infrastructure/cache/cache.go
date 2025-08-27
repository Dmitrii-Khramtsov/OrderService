// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/cache.go
package cache

import (
	"context"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
)

type Cache interface {
	Set(orderID string, order entities.Order)
	Get(orderID string) (entities.Order, bool)
	GetAll(limit int) ([]entities.Order, error)
	Delete(orderID string) bool
	Clear()
	Shutdown(ctx context.Context) error
}