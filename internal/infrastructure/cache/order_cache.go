// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/order_cache.go
package cache

import (
	"fmt"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"sync"
)

type Cache interface {
	Set(orderID string, order entities.Order)
	Get(orderID string) (entities.Order, bool)
	GetAll() []entities.Order
	Delete(orderID string) bool
	Clear()
}

type orderCache struct {
	sync.RWMutex
	data map[string]entities.Order
}

func NewOrderCache() Cache {
	return &orderCache{
		data: make(map[string]entities.Order, 1000),
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
func (c *orderCache) GetAll() []entities.Order {
	c.RLock()
	defer c.RUnlock()
	orders := make([]entities.Order, 0, len(c.data))
	for _, ord := range c.data {
		orders = append(orders, ord)
	}
	return orders
}

func (c *orderCache) Delete(orderID string) bool {
	c.Lock()
	defer c.Unlock()

	_, exist := c.data[orderID]
	if exist {
		delete(c.data, orderID)
		fmt.Printf("заказ с таким ID удалён: %v\n", orderID)
		return true
	}

	fmt.Printf("заказ с таким ID не удалён: %v\n", orderID)
	return false
}

func (c *orderCache) Clear() {
	c.Lock()
	defer c.Unlock()
	c.data = make(map[string]entities.Order, 1000)
}
