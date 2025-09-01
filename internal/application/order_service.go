// github.com/Dmitrii-Khramtsov/orderservice/internal/application/order_service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

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
	const op = "OrderService.SaveOrder"
	startTime := time.Now()
	defer s.logSaveDuration(order.OrderUID, startTime)

	if err := s.validateOrder(order); err != nil {
		return "", NewAppError(ErrCodeValidation, "order validation failed", op, err)
	}

	result := s.determineOrderResult(order)

	if err := s.saveToRepo(ctx, order); err != nil {
		return "", NewAppError(ErrCodeOrderSaveFailed, "failed to save order", op, err)
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
	const op = "OrderService.saveToRepo"
	
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		s.logger.Error("failed to save order to db",
			"order_id", order.OrderUID,
			"error", err,
		)
		if errors.Is(err, domain.ErrOrderNotFound) {
			return domain.ErrOrderNotFound
		}
		return NewAppError(ErrCodeOrderSaveFailed, "failed to save order to repository", op, err)
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
	const op = "OrderService.GetOrder"
	
	if order, found := s.cache.Get(id); found {
		s.logger.Info("order retrieved from cache", "order_id", id)
		return order, nil
	}

	dbOrder, err := s.fetchFromRepo(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			s.logger.Warn("order not found in db",
				"order_id", id,
				"error", err,
			)
			return entities.Order{}, domain.ErrOrderNotFound
		}
		return entities.Order{}, NewAppError(ErrCodeOrderReadFailed, "failed to retrieve order", op, err)
	}

	s.cache.Set(dbOrder.OrderUID, dbOrder)
	s.logger.Info("order retrieved from db and cached", "order_id", id)
	return dbOrder, nil
}

func (s *orderService) fetchFromRepo(ctx context.Context, id string) (entities.Order, error) {
	const op = "OrderService.fetchFromRepo"
	
	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return entities.Order{}, domain.ErrOrderNotFound
		}
		return entities.Order{}, NewAppError(ErrCodeOrderReadFailed, "failed to fetch order from repository", op, err)
	}
	return order, nil
}

func (s *orderService) DeleteOrder(ctx context.Context, id string) error {
	const op = "OrderService.DeleteOrder"
	
	if err := s.repo.DeleteOrder(ctx, id); err != nil {
		s.logger.Error("failed to delete order from db",
			"order_id", id,
			"error", err,
		)
		if errors.Is(err, domain.ErrOrderNotFound) {
			return domain.ErrOrderNotFound
		}
		return NewAppError(ErrCodeOrderDeleteFailed, "failed to delete order", op, err)
	}

	deleted := s.cache.Delete(id)
	if !deleted {
		s.logger.Warn("order not found in cache during deletion", "order_id", id)
	}

	s.logger.Info("order deleted", "order_id", id)
	return nil
}

func (s *orderService) ClearOrders(ctx context.Context) error {
	const op = "OrderService.ClearOrders"
	
	s.cache.Clear()

	if err := s.repo.ClearOrders(ctx); err != nil {
		s.logger.Error("failed to clear orders from db", "error", err)
		return NewAppError(ErrCodeOrderDeleteFailed, "failed to clear orders", op, err)
	}

	s.logger.Info("all orders cleared")
	return nil
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]entities.Order, error) {
	const op = "OrderService.GetAllOrders"
	
	ordersFromCache := s.cache.GetAll(s.getAllLimit)

	var orders []entities.Order
	var source string

	if len(ordersFromCache) > 0 {
		orders = ordersFromCache
		source = "cache"
	} else {
		s.logger.Info("cache is empty, retrieving orders from database")

		var err error
		orders, err = s.repo.GetAllOrders(ctx, s.getAllLimit, 0)
		if err != nil {
			s.logger.Error("failed to retrieve orders from database", "error", err)
			return nil, NewAppError(ErrCodeOrdersReadFailed, "failed to retrieve orders from database", op, err)
		}

		s.logger.Info("populating cache with orders from database", "count", len(orders))
		for _, order := range orders {
			s.cache.Set(order.OrderUID, order)
		}
		source = "database"
	}

	s.logger.Info("retrieved orders",
		"source", source,
		"count", len(orders))

	return orders, nil
}