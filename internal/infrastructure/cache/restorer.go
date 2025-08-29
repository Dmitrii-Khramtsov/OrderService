// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/restorer.go
package cache

import (
	"context"
	"time"

	"go.uber.org/zap"

	repo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

type CacheRestorer struct {
	cache     Cache
	repo      repo.OrderRepository
	logger    logger.LoggerInterface
	timeout   time.Duration
	batchSize int
}

func NewCacheRestorer(cache Cache, repo repo.OrderRepository, logger logger.LoggerInterface, timeout time.Duration, batchSize int) *CacheRestorer {
	return &CacheRestorer{
		cache:     cache,
		repo:      repo,
		logger:    logger,
		timeout:   timeout,
		batchSize: batchSize,
	}
}

func (r *CacheRestorer) Restore(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	total, err := r.repo.GetOrdersCount(ctx)
	if err != nil {
		r.logger.Error("failed to get orders count", zap.Error(err))
		return err
	}

	r.logger.Info("starting cache restoration", zap.Int("total_orders", total))

	var restored int
	for offset := 0; offset < total; offset += r.batchSize {
		select {
		case <-ctx.Done():
			r.logger.Warn("cache restoration timed out",
				zap.Int("restored", restored),
				zap.Int("total", total),
			)
			return ctx.Err()
		default:
			batch, err := r.repo.GetAllOrders(ctx, r.batchSize, offset)
			if err != nil {
				r.logger.Error("failed to get orders batch",
					zap.Error(err),
					zap.Int("offset", offset),
					zap.Int("limit", r.batchSize),
				)
				continue
			}

			for _, order := range batch {
				r.cache.Set(order.OrderUID, order)
				restored++
			}

			r.logger.Debug("processed batch",
				zap.Int("batch_size", len(batch)),
				zap.Int("restored", restored),
				zap.Int("progress", (restored*100)/total),
			)
		}
	}

	r.logger.Info("cache restoration completed",
		zap.Int("restored", restored),
		zap.Int("total", total),
	)
	return nil
}
