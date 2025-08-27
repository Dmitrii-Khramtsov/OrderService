// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/order_cache.go
package cache

import (
	"context"
	"sync"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// type Cache interface {
// 	Set(orderID string, order entities.Order)
// 	Get(orderID string) (entities.Order, bool)
// 	GetAll() ([]entities.Order, error)
// 	Delete(orderID string) bool
// 	Clear()
// 	Shutdown(ctx context.Context) error
// }

type orderCache struct {
	sync.RWMutex
	data map[string]entities.Order
	logger logger.LoggerInterface
}

func NewOrderCache(l logger.LoggerInterface) Cache {
	return &orderCache{
		data: make(map[string]entities.Order, 1000),
		logger: l,
	}
}

func (c *orderCache) Set(orderID string, order entities.Order) {
	c.Lock()
	defer c.Unlock()
	c.data[orderID] = order
}

func (c *orderCache) Get(orderID string) (entities.Order, bool) {
	c.RLock()
	defer c.RUnlock()
	order, exist := c.data[orderID]
	return order, exist
}

func (c *orderCache) GetAll() ([]entities.Order, error) {
	c.RLock()
	defer c.RUnlock()
	orders := make([]entities.Order, 0, len(c.data))
	for _, ord := range c.data {
		orders = append(orders, ord)
	}
	return orders, nil
}

func (c *orderCache) Delete(orderID string) bool {
	c.Lock()
	defer c.Unlock()

	_, exist := c.data[orderID]
	if exist {
		delete(c.data, orderID)
		c.logger.Info("order deleted", zap.String("order_id", orderID))
		return true
	}

	c.logger.Info("order not deleted", zap.String("order_id", orderID))
	return false
}

func (c *orderCache) Clear() {
	c.Lock()
	defer c.Unlock()
	c.data = make(map[string]entities.Order, 1000)
	c.logger.Info("cache cleared")
}

func (c *orderCache) Shutdown(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()


	c.data = make(map[string]entities.Order, 1000)

	c.logger.Info("cache cleared during shutdown")
	return nil
}
