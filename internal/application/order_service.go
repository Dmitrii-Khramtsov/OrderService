// github.com/Dmitrii-Khramtsov/orderservice/internal/application/order_service.go
package application

import (
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
)

type OrderServiceInterface interface {
	AddOrUpdateOrder(order entities.Order) (string, bool)
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

// func (s *orderService) AddOrder(order entities.Order) {
// 	s.cache.Set(order.OrderUID, order)
// }

func (s *orderService) AddOrUpdateOrder(order entities.Order) (string, bool) {
	existing, found := s.cache.Get(order.OrderUID)

	if !found {
		s.cache.Set(order.OrderUID, order)
		return "created", true
	}

	if existing.Equal(order) {
		return "exists", true
	}

	s.cache.Set(order.OrderUID, order)
	return "update", true
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
