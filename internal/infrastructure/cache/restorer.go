// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/restorer.go
package cache

import (
	"context"
	"time"

	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

type CacheRestorer struct {
	cache     domainrepo.Cache
	repo      domainrepo.OrderRepository
	logger    domainrepo.Logger
	timeout   time.Duration
	batchSize int
}

func NewCacheRestorer(cache domainrepo.Cache, repo domainrepo.OrderRepository, logger domainrepo.Logger, timeout time.Duration, batchSize int) *CacheRestorer {
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
		r.logger.Error("failed to get orders count", "error", err)
		return err
	}

	r.logger.Info("starting cache restoration", "total_orders", total)

	var restored int
	for offset := 0; offset < total; offset += r.batchSize {
		select {
		case <-ctx.Done():
			r.logger.Warn("cache restoration timed out",
				"restored", restored,
				"total", total,
			)
			return ctx.Err()
		default:
			batch, err := r.repo.GetAllOrders(ctx, r.batchSize, offset)
			if err != nil {
				r.logger.Error("failed to get orders batch",
					"error", err,
					"offset", offset,
					"limit", r.batchSize,
				)
				continue
			}

			for _, order := range batch {
				r.cache.Set(order.OrderUID, order)
				restored++
			}

			// процент восстановленных записей
			var progress int
			if total > 0 {
				progress = (restored * 100) / total
			}

			r.logger.Debug("processed batch",
				"batch_size", len(batch),
				"restored", restored,
				"progress", progress,
			)
		}
	}

	r.logger.Info("cache restoration completed",
		"restored", restored,
		"total", total,
	)
	return nil
}
