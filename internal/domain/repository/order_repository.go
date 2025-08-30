// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository/repository.go
package repository

import (
	"context"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order entities.Order) error
	GetOrder(ctx context.Context, id string) (entities.Order, error)
	GetAllOrders(ctx context.Context, limit, offset int) ([]entities.Order, error)
	GetOrdersCount(ctx context.Context) (int, error)
	DeleteOrder(ctx context.Context, id string) error
	ClearOrders(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
