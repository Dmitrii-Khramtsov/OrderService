// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/order_cache_test.go
package cache

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
)

type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Sync() error                             { return nil }
func (m *MockLogger) Shutdown(ctx context.Context) error      { return nil }

func TestOrderCache_SetAndGet(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 10)
	order := entities.Order{OrderUID: "test_order"}
	cache.Set("a", order)
	got, exists := cache.Get("a")
	if !exists {
		t.Fatal("Expected order to exist in cache")
	}
	if got.OrderUID != order.OrderUID {
		t.Errorf("Expected order %v, got %v", order, got)
	}
}

func TestOrderCache_GetMissing(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 10)
	_, exists := cache.Get("missing")
	if exists {
		t.Error("Expected missing order to not exist in cache")
	}
}

func TestOrderCache_Delete(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 10)
	order := entities.Order{OrderUID: "test_order"}
	cache.Set("a", order)
	deleted := cache.Delete("a")
	if !deleted {
		t.Error("Expected order to be deleted")
	}
	_, exists := cache.Get("a")
	if exists {
		t.Error("Expected deleted order to not exist in cache")
	}
}

func TestOrderCache_Clear(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 10)
	order1 := entities.Order{OrderUID: "test_order1"}
	order2 := entities.Order{OrderUID: "test_order2"}
	cache.Set("a", order1)
	cache.Set("b", order2)
	cache.Clear()
	_, exists1 := cache.Get("a")
	_, exists2 := cache.Get("b")
	if exists1 || exists2 {
		t.Error("Expected cache to be empty after Clear")
	}
}

func TestOrderCache_Update(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 10)
	order := entities.Order{OrderUID: "test_order"}
	cache.Set("a", order)
	updatedOrder := entities.Order{OrderUID: "updated_test_order"}
	cache.Set("a", updatedOrder)
	got, exists := cache.Get("a")
	if !exists {
		t.Fatal("Expected updated order to exist in cache")
	}
	if got.OrderUID != updatedOrder.OrderUID {
		t.Errorf("Expected updated order %v, got %v", updatedOrder, got)
	}
}

func TestOrderCache_ParallelAccess(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 100)
	order := entities.Order{OrderUID: "parallel_test_order"}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := strconv.Itoa(i)
			cache.Set(key, order)
			_, _ = cache.Get(key)
		}(i)
	}
	wg.Wait()
}

func TestOrderCache_Shutdown(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 10)
	order := entities.Order{OrderUID: "test_order"}
	cache.Set("a", order)
	err := cache.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Expected no error during Shutdown, got %v", err)
	}
	_, exists := cache.Get("a")
	if exists {
		t.Error("Expected cache to be empty after Shutdown")
	}
}

func TestOrderCache_GetAll(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 10)
	order1 := entities.Order{OrderUID: "test_order1"}
	order2 := entities.Order{OrderUID: "test_order2"}
	cache.Set("a", order1)
	cache.Set("b", order2)

	orders, err := cache.GetAll(10)
	if err != nil {
		t.Errorf("Expected no error during GetAll, got %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}
}

func TestOrderCache_UnlimitedCapacity(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 0)
	order1 := entities.Order{OrderUID: "test_order1"}
	order2 := entities.Order{OrderUID: "test_order2"}
	order3 := entities.Order{OrderUID: "test_order3"}

	cache.Set("a", order1)
	cache.Set("b", order2)
	cache.Set("c", order3)

	_, exists1 := cache.Get("a")
	_, exists2 := cache.Get("b")
	_, exists3 := cache.Get("c")

	if !exists1 || !exists2 || !exists3 {
		t.Error("Expected all orders to exist in unlimited cache")
	}
}

func TestOrderCache_Eviction(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 2)

	orderA := entities.Order{OrderUID: "orderA"}
	orderB := entities.Order{OrderUID: "orderB"}
	orderC := entities.Order{OrderUID: "orderC"}

	cache.Set("a", orderA)
	cache.Set("b", orderB)
	cache.Set("c", orderC)

	_, existsA := cache.Get("a")
	if existsA {
		t.Error("Expected order 'a' to be evicted from cache")
	}

	_, existsB := cache.Get("b")
	_, existsC := cache.Get("c")
	if !existsB || !existsC {
		t.Error("Expected orders 'b' and 'c' to exist in cache")
	}
}

func TestOrderCache_FreshnessUpdate(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 2)

	orderA := entities.Order{OrderUID: "orderA"}
	orderB := entities.Order{OrderUID: "orderB"}
	orderC := entities.Order{OrderUID: "orderC"}

	cache.Set("a", orderA)
	cache.Set("b", orderB)
	cache.Get("a")
	cache.Set("c", orderC)

	_, existsB := cache.Get("b")
	if existsB {
		t.Error("Expected order 'b' to be evicted from cache")
	}

	_, existsA := cache.Get("a")
	_, existsC := cache.Get("c")
	if !existsA || !existsC {
		t.Error("Expected orders 'a' and 'c' to exist in cache")
	}
}

func TestOrderCache_KeyOverwrite(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 2)

	orderV1 := entities.Order{OrderUID: "orderV1"}
	orderV2 := entities.Order{OrderUID: "orderV2"}

	cache.Set("a", orderV1)
	cache.Set("a", orderV2)

	got, exists := cache.Get("a")
	if !exists {
		t.Fatal("Expected order 'a' to exist in cache")
	}
	if got.OrderUID != orderV2.OrderUID {
		t.Errorf("Expected orderV2, got %v", got.OrderUID)
	}
}

func TestOrderCache_GetAllOrder(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 2)

	orderA := entities.Order{OrderUID: "orderA"}
	orderB := entities.Order{OrderUID: "orderB"}

	cache.Set("a", orderA)
	cache.Set("b", orderB)
	cache.Get("a")

	orders, err := cache.GetAll(2)
	if err != nil {
		t.Errorf("Expected no error during GetAll, got %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}

	if orders[0].OrderUID != "orderA" {
		t.Errorf("Expected first order to be 'orderA', got %v", orders[0].OrderUID)
	}
}

func TestOrderCache_CapacityOne(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderLRUCache(log, 1)

	orderA := entities.Order{OrderUID: "orderA"}
	orderB := entities.Order{OrderUID: "orderB"}

	cache.Set("a", orderA)
	cache.Set("b", orderB)

	_, existsA := cache.Get("a")
	if existsA {
		t.Error("Expected order 'a' to be evicted from cache")
	}

	_, existsB := cache.Get("b")
	if !existsB {
		t.Error("Expected order 'b' to exist in cache")
	}
}
