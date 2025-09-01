// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository/cache.go
package repository

import (
	"context"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
)

type Cache interface {
	Get(key string) (entities.Order, bool)
	Set(key string, order entities.Order)
	Delete(key string) bool
	Clear()
	GetAll(limit int) []entities.Order
	Shutdown(ctx context.Context) error
}
