// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/cache.go
package factory

import (
	"time"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	repo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func NewCache(l logger.LoggerInterface, capacity int) cache.Cache {
	return cache.NewOrderLRUCache(l, capacity)
}

func NewCacheRestorer(c cache.Cache, r repo.OrderRepository, l logger.LoggerInterface) *cache.CacheRestorer {
	return cache.NewCacheRestorer(
		c, 
		r, 
		l, 
		5*time.Minute, // timeout
		1000,          // batch size
	)
}