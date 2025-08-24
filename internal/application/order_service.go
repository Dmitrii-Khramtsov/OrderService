// github.com/Dmitrii-Khramtsov/orderservice/internal/application/order_service.go
package application

import (
	"fmt"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"go.uber.org/zap"
)

type OrderServiceInterface interface {
	SaveOrder(order entities.Order) (OrderResult, error)
	GetOrder(id string) (entities.Order, error)
	GetAllOrder() ([]entities.Order, error)
	DelOrder(id string) error
	ClearOrder()
}

type orderService struct {
	cache cache.Cache
	logger logger.LoggerInterface
	// db    repositories.Repository
}

func NewOrderService(c cache.Cache, l logger.LoggerInterface) OrderServiceInterface {
	return &orderService{
		cache: c,
		logger: l,
	}
}

type OrderResult string

const (
	OrderCreated OrderResult = "created"
	OrderUpdated OrderResult = "updated"
	OrderExists  OrderResult = "exists"
)

func (s *orderService) SaveOrder(order entities.Order) (OrderResult, error) {
	if err := order.Validate(); err != nil {
		s.logger.Warn("order validation failed",
		zap.String("order_id", order.OrderUID),
		zap.Error(err),
	)
		return "", fmt.Errorf("%w: %v", domain.ErrInvalidOrder, err)
	}

	existing, found := s.cache.Get(order.OrderUID)

	if !found {
		s.cache.Set(order.OrderUID, order)
		s.logger.Info("order created", zap.String("order_id", order.OrderUID))
		return OrderCreated, nil
	}

	if existing.Equal(order) {
		s.logger.Info("order already exists", zap.String("order_id", order.OrderUID))
		return OrderExists, nil
	}

	s.cache.Set(order.OrderUID, order)
	s.logger.Info("order updated", zap.String("order_id", order.OrderUID))
	return OrderUpdated, nil
}

func (s *orderService) GetOrder(id string) (entities.Order, error) {
	order, found := s.cache.Get(id)
	if !found {
		s.logger.Warn("order not found", zap.String("order_id", id))
		return entities.Order{}, domain.ErrOrderNotFound
	}
	return order, nil
}

func (s *orderService) GetAllOrder() ([]entities.Order, error) {
	return s.cache.GetAll()
}

func (s *orderService) DelOrder(id string) error {
	ok := s.cache.Delete(id)
	if !ok {
		s.logger.Warn("delete failed, order not found", zap.String("order_id", id))
		return domain.ErrOrderNotFound
	}

	s.logger.Info("order deleted", zap.String("order_id", id))
	return nil
}

func (s *orderService) ClearOrder() {
	s.cache.Clear()
}
