// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database/retrying_repository.go
package repository

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

type RetryingOrderRepository struct {
	repo   OrderRepository
	logger logger.LoggerInterface
	config *RetryConfig
}

type RetryConfig struct {
	MaxElapsedTime      time.Duration
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
}

func NewRetryingOrderRepository(repo OrderRepository, logger logger.LoggerInterface, config *RetryConfig) *RetryingOrderRepository {
	return &RetryingOrderRepository{
		repo:   repo,
		logger: logger,
		config: config,
	}
}

func (r *RetryingOrderRepository) withRetry(ctx context.Context, operation func() error) error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = r.config.MaxElapsedTime
	expBackoff.InitialInterval = r.config.InitialInterval
	expBackoff.RandomizationFactor = r.config.RandomizationFactor
	expBackoff.Multiplier = r.config.Multiplier
	expBackoff.MaxInterval = r.config.MaxInterval

	return backoff.Retry(func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(ctx.Err())
		default:
			return operation()
		}
	}, expBackoff)
}

func (r *RetryingOrderRepository) SaveOrder(ctx context.Context, order entities.Order) error {
	return r.withRetry(ctx, func() error {
		err := r.repo.SaveOrder(ctx, order)
		if err != nil {
			r.logger.Warn("failed to save order, retrying",
				zap.String("order_uid", order.OrderUID),
				zap.Error(err),
			)
		}
		return err
	})
}

func (r *RetryingOrderRepository) GetOrder(ctx context.Context, id string) (entities.Order, error) {
	var order entities.Order
	var err error

	operation := func() error {
		order, err = r.repo.GetOrder(ctx, id)
		if err != nil {
			r.logger.Warn("failed to get order, retrying",
				zap.String("order_uid", id),
				zap.Error(err),
			)
		}
		return err
	}

	err = r.withRetry(ctx, operation)
	return order, err
}

func (r *RetryingOrderRepository) GetAllOrders(ctx context.Context, limit, offset int) ([]entities.Order, error) {
	var orders []entities.Order
	var err error

	operation := func() error {
		orders, err = r.repo.GetAllOrders(ctx, limit, offset)
		if err != nil {
			r.logger.Warn("failed to get all orders, retrying", zap.Error(err))
		}
		return err
	}

	err = r.withRetry(ctx, operation)
	return orders, err
}

func (r *RetryingOrderRepository) GetOrdersCount(ctx context.Context) (int, error) {
	var count int
	var err error

	operation := func() error {
		count, err = r.repo.GetOrdersCount(ctx)
		if err != nil {
			r.logger.Warn("failed to get orders count, retrying", zap.Error(err))
		}
		return err
	}

	err = r.withRetry(ctx, operation)
	return count, err
}

func (r *RetryingOrderRepository) DeleteOrder(ctx context.Context, id string) error {
	return r.withRetry(ctx, func() error {
		err := r.repo.DeleteOrder(ctx, id)
		if err != nil {
			r.logger.Warn("failed to delete order, retrying",
				zap.String("order_uid", id),
				zap.Error(err),
			)
		}
		return err
	})
}

func (r *RetryingOrderRepository) ClearOrders(ctx context.Context) error {
	return r.withRetry(ctx, func() error {
		err := r.repo.ClearOrders(ctx)
		if err != nil {
			r.logger.Warn("failed to clear orders, retrying", zap.Error(err))
		}
		return err
	})
}

func (r *RetryingOrderRepository) Shutdown(ctx context.Context) error {
	return r.repo.Shutdown(ctx)
}
