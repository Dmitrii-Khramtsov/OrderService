// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/order_cache_lru.go
package cache

import (
	"container/list"
	"context"
	"sync"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"go.uber.org/zap"
)

type entry struct {
	kay   string
	value entities.Order
}

type orderLRUCache struct {
	sync.RWMutex
	capacity int
	cache    map[string]*list.Element
	ll       *list.List
	logger   logger.LoggerInterface
}

func NewOrderLRUCache(l logger.LoggerInterface, capacity int) Cache {
	return &orderLRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element, 1000),
		ll:       list.New(),
		logger:   l,
	}
}

func (c *orderLRUCache) Set(orderID string, order entities.Order) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	if elem, exist := c.cache[orderID]; exist {
		entry := elem.Value.(*entry)
		entry.value = order
		c.ll.MoveToFront(elem)
		c.logger.Info("order updated in cache", zap.String("order_id", orderID))
		return
	}

	if c.ll.Len() >= c.capacity {
		lastElem := c.ll.Back()
		if lastElem != nil {
			lastEntry := lastElem.Value.(*entry)
			delete(c.cache, lastEntry.kay)
			c.ll.Remove(lastElem)
			c.logger.Info("cache execeeded, most unused order deleted", zap.String("order_id", lastEntry.kay))
		}
	}

	newEntry := &entry{kay: orderID, value: order}
	newElem := c.ll.PushFront(newEntry)
	c.cache[orderID] = newElem
}

func (c *orderLRUCache) Get(orderID string) (entities.Order, bool) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()

	if elem, exist := c.cache[orderID]; exist {
		c.ll.MoveToFront(elem)
		entry := elem.Value.(*entry)
		c.logger.Info("retrieved order from cache", zap.String("order_id", orderID))
		return entry.value, true
	}

	c.logger.Info("there is no such order", zap.String("order_id", orderID))
	return entities.Order{}, false
}

func (c *orderLRUCache) GetAll(limit int) ([]entities.Order, error) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()

	if limit <= 0 {
		c.logger.Info("limit is less than or equal to zero, returning empty slice", zap.Int("limit", limit))
		return []entities.Order{}, nil
	}

	capacity := limit
	if len(c.cache) < limit {
		capacity = len(c.cache)
		c.logger.Info("limit exceeds cache size, adjusting capacity",
			zap.Int("limit", limit),
			zap.Int("cache_size", len(c.cache)),
			zap.Int("adjusted_capacity", capacity))
	}

	orders := make([]entities.Order, 0, capacity)
	count := 0
	for elem := c.ll.Front(); elem != nil && count < limit; elem = elem.Next() {
		entry := elem.Value.(*entry)
		orders = append(orders, entry.value)
		count++
	}

	c.logger.Info("Retrieved orders from cache",
		zap.Int("requested_limit", limit),
		zap.Int("returned_count", count))

	return orders, nil
}

func (c *orderLRUCache) Delete(orderID string) bool {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	if elem, exist := c.cache[orderID]; exist {
		c.ll.Remove(elem)
		delete(c.cache, orderID)
		c.logger.Info("order deleted", zap.String("order_id", orderID))
		return true
	}

	c.logger.Info("order not deleted", zap.String("order_id", orderID))
	return false
}

func (c *orderLRUCache) Clear() {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	c.cache = make(map[string]*list.Element, 1000)
	c.ll.Init()
	c.logger.Info("cache cleared")
}

func (c *orderLRUCache) Shutdown(ctx context.Context) error {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	c.cache = make(map[string]*list.Element, 1000)
	c.ll.Init()
	
	c.logger.Info("cache cleared during shutdown")
	return nil
}
