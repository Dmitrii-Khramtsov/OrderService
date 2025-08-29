package factory

import (
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func NewCache(l logger.LoggerInterface, capacity int) cache.Cache {
	return cache.NewOrderLRUCache(l, capacity)
}
