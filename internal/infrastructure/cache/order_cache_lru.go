// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/order_cache_lru.go
package cache

import (
	"container/list"
	"context"
	"sync"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

type entry struct {
	key   string
	value entities.Order
}

type orderLRUCache struct {
	sync.RWMutex
	capacity int
	cache    map[string]*list.Element
	ll       *list.List
	logger   domainrepo.Logger
}

func NewOrderLRUCache(l domainrepo.Logger, capacity int) domainrepo.Cache {
	if capacity <= 0 {
		capacity = -1
		l.Info("cache capacity is set to unlimited")
	}
	return &orderLRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element, max(0, capacity)),
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
		c.logger.Debug("order updated in cache", "order_id", orderID)
		return
	}

	if c.capacity > 0 && c.ll.Len() >= c.capacity {
		lastElem := c.ll.Back()
		if lastElem != nil {
			lastEntry := lastElem.Value.(*entry)
			delete(c.cache, lastEntry.key)
			c.ll.Remove(lastElem)
			c.logger.Debug("cache exceeded, most unused order deleted", "order_id", lastEntry.key)
		}
	}

	newEntry := &entry{key: orderID, value: order}
	newElem := c.ll.PushFront(newEntry)
	c.cache[orderID] = newElem
}

// не стал использовать RLock и оборачивать отдельно Lock - c.ll.MoveToFront(elem)
func (c *orderLRUCache) Get(orderID string) (entities.Order, bool) {
	c.Lock()
	defer c.Unlock()

	elem, exist := c.cache[orderID]
	if !exist {
		c.logger.Debug("there is no such order", "order_id", orderID)
		return entities.Order{}, false
	}

	c.ll.MoveToFront(elem)
	entry := elem.Value.(*entry)
	c.logger.Debug("retrieved order from cache", "order_id", orderID)
	return entry.value, true
}

func (c *orderLRUCache) GetAll(limit int) []entities.Order {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()

	if limit <= 0 {
		c.logger.Debug("limit is less than or equal to zero, returning empty slice", "limit", limit)
		return []entities.Order{}
	}

	capacity := limit
	if len(c.cache) < limit {
		capacity = len(c.cache)
		c.logger.Debug("limit exceeds cache size, adjusting capacity",
			"limit", limit,
			"cache_size", len(c.cache),
			"adjusted_capacity", capacity)
	}

	orders := make([]entities.Order, 0, capacity)
	count := 0
	for elem := c.ll.Front(); elem != nil && count < limit; elem = elem.Next() {
		entry := elem.Value.(*entry)
		orders = append(orders, entry.value)
		count++
	}

	c.logger.Debug("Retrieved orders from cache",
		"requested_limit", limit,
		"returned_count", count)

	return orders
}

func (c *orderLRUCache) Delete(orderID string) bool {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	if elem, exist := c.cache[orderID]; exist {
		c.ll.Remove(elem)
		delete(c.cache, orderID)
		c.logger.Info("order deleted", "order_id", orderID)
		return true
	}

	c.logger.Info("order not deleted", "order_id", orderID)
	return false
}

func (c *orderLRUCache) reset() {
	cacheCapacity := c.capacity
	if cacheCapacity <= 0 {
		cacheCapacity = 0
	}
	c.ll.Init()
	c.cache = make(map[string]*list.Element, cacheCapacity)
}

func (c *orderLRUCache) Clear() {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.reset()
	c.logger.Info("cache cleared")
}

func (c *orderLRUCache) Shutdown(ctx context.Context) error {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.reset()
	c.logger.Info("cache cleared during shutdown")
	return nil
}
