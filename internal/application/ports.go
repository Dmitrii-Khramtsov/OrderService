// github.com/Dmitrii-Khramtsov/orderservice/internal/application/order_service.go
package application

import (
	"context"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
)

type OrderServiceInterface interface {
	SaveOrder(ctx context.Context, order entities.Order) (OrderResult, error)
	GetOrder(ctx context.Context, id string) (entities.Order, error)
	GetAllOrders(ctx context.Context) ([]entities.Order, error)
	DeleteOrder(ctx context.Context, id string) error
	ClearOrders(ctx context.Context) error
}
