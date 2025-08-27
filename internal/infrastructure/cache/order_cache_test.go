// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache/order_cache_test.go
package cache

import (
	"context"
	"sync"
	"testing"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"go.uber.org/zap"
)

type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {}
func (m *MockLogger) Warn(msg string, fields ...zap.Field)  {}
func (m *MockLogger) Error(msg string, fields ...zap.Field) {}
func (m *MockLogger) Sync()                                   {}
func (m *MockLogger) Shutdown(ctx context.Context) error     { return nil }

func TestOrderCache_SetAndGet(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderCache(log)

	order := entities.Order{OrderUID: "test_order"}

	cache.Set("a", order)
	got, exists := cache.Get("a")

	if !exists {
		t.Fatal("Expected order to exist in cache")
	}
	if !got.Equal(order) {
		t.Errorf("Expected order %v, got %v", order, got)
	}
}

func TestOrderCache_GetMissing(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderCache(log)

	_, exists := cache.Get("missing")
	if exists {
		t.Error("Expected missing order to not exist in cache")
	}
}

func TestOrderCache_Delete(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderCache(log)

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
	cache := NewOrderCache(log)

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

func TestOrderCache_ParallelAccess(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderCache(log)

	order := entities.Order{OrderUID: "parallel_test_order"}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.Set(string(rune(i)), order)
			_, _ = cache.Get(string(rune(i)))
		}(i)
	}
	wg.Wait()
}

func TestOrderCache_Shutdown(t *testing.T) {
	log := &MockLogger{}
	cache := NewOrderCache(log)

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
