// github.com/Dmitrii-Khramtsov/orderservice/internal/application/order_service.go
package application

import (
	"context"
	"fmt"
	"time"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

type OrderServiceInterface interface {
	SaveOrder(ctx context.Context, order entities.Order) (OrderResult, error)
	GetOrder(ctx context.Context, id string) (entities.Order, error)
	GetAllOrder(ctx context.Context) ([]entities.Order, error)
	DelOrder(ctx context.Context, id string) error
	ClearOrder(ctx context.Context) error
}

type orderService struct {
	cache       domainrepo.Cache
	logger      domainrepo.Logger
	repo        domainrepo.OrderRepository
	getAllLimit int
}

func NewOrderService(c domainrepo.Cache, l domainrepo.Logger, r domainrepo.OrderRepository, limit int) OrderServiceInterface {
	return &orderService{
		cache:       c,
		logger:      l,
		repo:        r,
		getAllLimit: limit,
	}
}

type OrderResult string

const (
	OrderCreated OrderResult = "created"
	OrderUpdated OrderResult = "updated"
	OrderExists  OrderResult = "exists"
)

func (s *orderService) SaveOrder(ctx context.Context, order entities.Order) (OrderResult, error) {
	startTime := time.Now()
	defer s.logSaveDuration(order.OrderUID, startTime)

	if err := s.validateOrder(order); err != nil {
		return "", err
	}

	result := s.determineOrderResult(order)

	if err := s.saveToRepo(ctx, order); err != nil {
		return "", err
	}

	s.updateCache(order, result)

	return result, nil
}

func (s *orderService) validateOrder(order entities.Order) error {
	if err := order.Validate(); err != nil {
		s.logger.Warn("order validation failed",
			"order_id", order.OrderUID,
			"error", err,
		)
		return fmt.Errorf("%w: %v", domain.ErrInvalidOrder, err)
	}
	return nil
}

func (s *orderService) determineOrderResult(order entities.Order) OrderResult {
	existing, found := s.cache.Get(order.OrderUID)
	switch {
	case !found:
		return OrderCreated
	case existing.Equal(order):
		return OrderExists
	default:
		return OrderUpdated
	}
}

func (s *orderService) saveToRepo(ctx context.Context, order entities.Order) error {
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		s.logger.Error("failed to save order to db",
			"order_id", order.OrderUID,
			"error", err,
		)
		return err
	}
	return nil
}

func (s *orderService) updateCache(order entities.Order, result OrderResult) {
	s.cache.Set(order.OrderUID, order)
	s.logger.Info("order saved",
		"order_id", order.OrderUID,
		"result", string(result),
	)
}

func (s *orderService) logSaveDuration(orderID string, startTime time.Time) {
	s.logger.Info("SaveOrder completed",
		"order_id", orderID,
		"duration", time.Since(startTime),
	)
}

func (s *orderService) GetOrder(ctx context.Context, id string) (entities.Order, error) {
	if order, found := s.cache.Get(id); found {
		s.logger.Info("order retrieved from cache", "order_id", id)
		return order, nil
	}

	dbOrder, err := s.fetchFromRepo(ctx, id)
	if err != nil {
		s.logger.Warn("order not found in db",
			"order_id", id,
			"error", err,
		)
		return entities.Order{}, domain.ErrOrderNotFound
	}

	s.cache.Set(dbOrder.OrderUID, dbOrder)
	s.logger.Info("order retrieved from db and cached", "order_id", id)
	return dbOrder, nil
}

func (s *orderService) fetchFromRepo(ctx context.Context, id string) (entities.Order, error) {
	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return entities.Order{}, err
	}
	return order, nil
}

func (s *orderService) DelOrder(ctx context.Context, id string) error {
	if err := s.repo.DeleteOrder(ctx, id); err != nil {
		s.logger.Error("failed to delete order from db",
			"order_id", id,
			"error", err,
		)
		return err
	}

	ok := s.cache.Delete(id)
	if !ok {
		s.logger.Warn("order not found in cache", "order_id", id)
	}

	s.logger.Info("order deleted", "order_id", id)
	return nil
}

func (s *orderService) ClearOrder(ctx context.Context) error {
	s.cache.Clear()

	if err := s.repo.ClearOrders(ctx); err != nil {
		s.logger.Error("failed to clear orders from db", "error", err)
		return err
	}

	s.logger.Info("all orders cleared")
	return nil
}

func (s *orderService) GetAllOrder(ctx context.Context) ([]entities.Order, error) {
	orders, err := s.cache.GetAll(s.getAllLimit)
	if err != nil {
		s.logger.Error("failed to retrieve orders from cache", "error", err)
		return nil, fmt.Errorf("failed to retrieve orders from cache: %w", err)
	}
	s.logger.Info("retrieved all orders from cache", "count", len(orders))
	return orders, err
}
