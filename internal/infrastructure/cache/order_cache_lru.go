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

func newOrderCache(l logger.LoggerInterface, capacity int) Cache {
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
		c.logger.Info("Order updated in cache", zap.String("order_id", orderID))
		return
	}

	if c.ll.Len() >= c.capacity {
		lastElem := c.ll.Back()
		if lastElem != nil {
			lastEntry := lastElem.Value.(*entry)
			delete(c.cache, lastEntry.kay)
			c.ll.Remove(lastElem)
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
		c.logger.Info("Retrieved order from cache", zap.String("order_id", orderID))
		return entry.value, true
	}

	c.logger.Info("there is no such order", zap.String("order_id", orderID))
	return entities.Order{}, false
}

func (c *orderLRUCache) GetAll() ([]entities.Order, error)

func (c *orderLRUCache) Delete(orderID string) bool

func (c *orderLRUCache) Clear()

func (c *orderLRUCache) Shutdown(ctx context.Context) error
