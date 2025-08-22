// github.com/Dmitrii-Khramtsov/orderservice/internal/application/order_service.go
package application

import (
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
)

type OrderServiceInterface interface {
	SaveOrder(order entities.Order) OrderResult
	GetOrder(id string) (entities.Order, bool)
	GetAllOrder() []entities.Order
	DelOrder(id string) bool
	ClearOrder()
}

type orderService struct {
	cache cache.Cache
	// db    repositories.Repository
}

func NewOrderService(c cache.Cache) OrderServiceInterface {
	return &orderService{
		cache: c,
	}
}

type OrderResult string

const (
	OrderCreated OrderResult = "created"
	OrderUpdated OrderResult = "updated"
	OrderExists  OrderResult = "exists"
)

func (s *orderService) SaveOrder(order entities.Order) OrderResult {
	existing, found := s.cache.Get(order.OrderUID)

	if !found {
		s.cache.Set(order.OrderUID, order)
		return OrderCreated
	}

	if existing.Equal(order) {
		return OrderExists
	}

	s.cache.Set(order.OrderUID, order)
	return OrderUpdated
}

func (s *orderService) GetOrder(id string) (entities.Order, bool) {
	return s.cache.Get(id)
}

func (s *orderService) GetAllOrder() []entities.Order {
	return s.cache.GetAll()
}

func (s *orderService) DelOrder(id string) bool {
	return s.cache.Delete(id)
}

func (s *orderService) ClearOrder() {
	s.cache.Clear()
}
