// github.com/Dmitrii-Khramtsov/orderservice/internal/application/order_service_test.go
package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCache struct{ mock.Mock }
func (m *mockCache) Get(key string) (entities.Order, bool) {
	args := m.Called(key)
	return args.Get(0).(entities.Order), args.Bool(1)
}
func (m *mockCache) Set(key string, order entities.Order) { m.Called(key, order) }
func (m *mockCache) Delete(key string) bool { return m.Called(key).Bool(0) }
func (m *mockCache) Clear() { m.Called() }
func (m *mockCache) GetAll(limit int) ([]entities.Order, error) {
	args := m.Called(limit)
	return args.Get(0).([]entities.Order), args.Error(1)
}
func (m *mockCache) Shutdown(ctx context.Context) error { return m.Called(ctx).Error(0) }

type mockRepo struct{ mock.Mock }
func (m *mockRepo) SaveOrder(ctx context.Context, order entities.Order) error {
	return m.Called(ctx, order).Error(0)
}
func (m *mockRepo) GetOrder(ctx context.Context, id string) (entities.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entities.Order), args.Error(1)
}
func (m *mockRepo) GetAllOrders(ctx context.Context, limit, offset int) ([]entities.Order, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entities.Order), args.Error(1)
}
func (m *mockRepo) GetOrdersCount(ctx context.Context) (int, error) { args := m.Called(ctx); return args.Int(0), args.Error(1) }
func (m *mockRepo) DeleteOrder(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }
func (m *mockRepo) ClearOrders(ctx context.Context) error { return m.Called(ctx).Error(0) }
func (m *mockRepo) Shutdown(ctx context.Context) error { return m.Called(ctx).Error(0) }

type mockLogger struct{ mock.Mock }
func (m *mockLogger) Debug(msg string, fields ...interface{}) { m.Called(msg, fields) }
func (m *mockLogger) Info(msg string, fields ...interface{})  { m.Called(msg, fields) }
func (m *mockLogger) Warn(msg string, fields ...interface{})  { m.Called(msg, fields) }
func (m *mockLogger) Error(msg string, fields ...interface{}) { m.Called(msg, fields) }
func (m *mockLogger) Sync() error { return m.Called().Error(0) }
func (m *mockLogger) Shutdown(ctx context.Context) error { return m.Called(ctx).Error(0) }

func sampleOrder() entities.Order {
	return entities.Order{
		OrderUID:    "123",
		TrackNumber: "TRACK123",
		Items: []entities.Item{
			{ChrtID: 1, TrackNumber: "TRACK123", Price: 100, RID: "RID1", Name: "item1", TotalPrice: 100, NmID: 1},
		},
		Delivery: entities.Delivery{Name: "John Doe", Email: "test@example.com", Phone: "+123456"},
		Payment:  entities.Payment{Amount: 100},
	}
}

func TestSaveOrder_NewOrder(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockRepo)
	logger := new(mockLogger)

	order := sampleOrder()
	cache.On("Get", order.OrderUID).Return(entities.Order{}, false)
	cache.On("Set", order.OrderUID, order).Return()
	repo.On("SaveOrder", mock.Anything, order).Return(nil)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	s := application.NewOrderService(cache, logger, repo, 10)
	res, err := s.SaveOrder(context.Background(), order)

	assert.NoError(t, err)
	assert.Equal(t, application.OrderCreated, res)
	cache.AssertCalled(t, "Get", order.OrderUID)
	cache.AssertCalled(t, "Set", order.OrderUID, order)
	repo.AssertCalled(t, "SaveOrder", mock.Anything, order)
}

func TestGetOrder_FromCache(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockRepo)
	logger := new(mockLogger)

	order := sampleOrder()
	cache.On("Get", order.OrderUID).Return(order, true)
	logger.On("Info", mock.Anything, mock.Anything).Return()

	s := application.NewOrderService(cache, logger, repo, 10)
	got, err := s.GetOrder(context.Background(), order.OrderUID)

	assert.NoError(t, err)
	assert.Equal(t, order, got)
}

func TestGetOrder_FromRepo(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockRepo)
	logger := new(mockLogger)

	order := sampleOrder()
	cache.On("Get", order.OrderUID).Return(entities.Order{}, false)
	repo.On("GetOrder", mock.Anything, order.OrderUID).Return(order, nil)
	cache.On("Set", order.OrderUID, order).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	s := application.NewOrderService(cache, logger, repo, 10)
	got, err := s.GetOrder(context.Background(), order.OrderUID)

	assert.NoError(t, err)
	assert.Equal(t, order, got)
}

func TestDeleteOrder_Success(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockRepo)
	logger := new(mockLogger)

	id := "123"
	repo.On("DeleteOrder", mock.Anything, id).Return(nil)
	cache.On("Delete", id).Return(true)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	s := application.NewOrderService(cache, logger, repo, 10)
	err := s.DeleteOrder(context.Background(), id)

	assert.NoError(t, err)
}

func TestClearOrders_Success(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockRepo)
	logger := new(mockLogger)

	repo.On("ClearOrders", mock.Anything).Return(nil)
	cache.On("Clear").Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	s := application.NewOrderService(cache, logger, repo, 10)
	err := s.ClearOrders(context.Background())

	assert.NoError(t, err)
}

func TestGetAllOrders_FromCache(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockRepo)
	logger := new(mockLogger)

	orders := []entities.Order{sampleOrder()}
	cache.On("GetAll", 10).Return(orders, nil)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	s := application.NewOrderService(cache, logger, repo, 10)
	got, err := s.GetAllOrders(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, orders, got)
}

func TestGetAllOrders_FromRepo(t *testing.T) {
	cache := new(mockCache)
	repo := new(mockRepo)
	logger := new(mockLogger)

	orders := []entities.Order{sampleOrder()}
	cache.On("GetAll", 10).Return([]entities.Order{}, errors.New("cache miss"))
	repo.On("GetAllOrders", mock.Anything, 10, 0).Return(orders, nil)
	cache.On("Set", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	s := application.NewOrderService(cache, logger, repo, 10)
	got, err := s.GetAllOrders(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, orders, got)
}
