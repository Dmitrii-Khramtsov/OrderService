// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/cache.go
package factory

import (
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	infracache "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
)

func NewCache(l domainrepo.Logger, capacity int) domainrepo.Cache {
	return infracache.NewOrderLRUCache(l, capacity)
}

func NewCacheRestorer(cfg *config.Config, c domainrepo.Cache, r domainrepo.OrderRepository, l domainrepo.Logger) *infracache.CacheRestorer {
	return infracache.NewCacheRestorer(
		c,
		r,
		l,
		cfg.Cache.Restoration.Timeout,
		cfg.Cache.Restoration.BatchSize,
	)
}
