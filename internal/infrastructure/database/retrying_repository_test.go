// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database/retrying_repository_test.go
package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) SaveOrder(ctx context.Context, order entities.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrder(ctx context.Context, id string) (entities.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entities.Order), args.Error(1)
}

func (m *MockOrderRepository) GetAllOrders(ctx context.Context, limit, offset int) ([]entities.Order, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]entities.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockOrderRepository) DeleteOrder(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) ClearOrders(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockOrderRepository) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRetryingOrderRepository_Success(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	logger, _ := logger.NewLogger(logger.DEV)
	
	retryConfig := &RetryConfig{
		MaxElapsedTime:     1 * time.Second,
		InitialInterval:    10 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:         1.5,
		MaxInterval:        100 * time.Millisecond,
	}
	
	repo := NewRetryingOrderRepository(mockRepo, logger, retryConfig)
	
	testOrder := entities.Order{OrderUID: "test123"}
	mockRepo.On("SaveOrder", mock.Anything, testOrder).Return(nil).Once()
	
	ctx := context.Background()
	err := repo.SaveOrder(ctx, testOrder)
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRetryingOrderRepository_WithRetries(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	logger, _ := logger.NewLogger(logger.DEV)
	
	retryConfig := &RetryConfig{
		MaxElapsedTime:     5 * time.Second,
		InitialInterval:    10 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:         1.5,
		MaxInterval:        100 * time.Millisecond,
	}
	
	repo := NewRetryingOrderRepository(mockRepo, logger, retryConfig)
	
	testOrder := entities.Order{OrderUID: "test123"}
	
	mockRepo.On("SaveOrder", mock.Anything, testOrder).Return(errors.New("temporary error")).Twice()
	mockRepo.On("SaveOrder", mock.Anything, testOrder).Return(nil).Once()
	
	ctx := context.Background()
	err := repo.SaveOrder(ctx, testOrder)
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	mockRepo.AssertNumberOfCalls(t, "SaveOrder", 3)
}

func TestRetryingOrderRepository_MaxRetriesExceeded(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	logger, _ := logger.NewLogger(logger.DEV)
	
	retryConfig := &RetryConfig{
		MaxElapsedTime:     50 * time.Millisecond,
		InitialInterval:    1 * time.Millisecond,
		RandomizationFactor: 0.1,
		Multiplier:         1.1,
		MaxInterval:        5 * time.Millisecond,
	}
	
	repo := NewRetryingOrderRepository(mockRepo, logger, retryConfig)
	
	testOrder := entities.Order{OrderUID: "test123"}
	
	mockRepo.On("SaveOrder", mock.Anything, testOrder).Return(errors.New("permanent error"))
	
	ctx := context.Background()
	err := repo.SaveOrder(ctx, testOrder)
	
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	
	calls := len(mockRepo.Calls)
	assert.Greater(t, calls, 1, "Expected more than 1 retry attempt, got %d", calls)
}

func TestRetryingOrderRepository_GetOrderWithRetry(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	logger, _ := logger.NewLogger(logger.DEV)
	
	retryConfig := &RetryConfig{
		MaxElapsedTime:     5 * time.Second,
		InitialInterval:    10 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:         1.5,
		MaxInterval:        100 * time.Millisecond,
	}
	
	repo := NewRetryingOrderRepository(mockRepo, logger, retryConfig)
	
	testOrder := entities.Order{OrderUID: "test123"}
	
	mockRepo.On("GetOrder", mock.Anything, "test123").Return(entities.Order{}, errors.New("temporary error")).Once()
	mockRepo.On("GetOrder", mock.Anything, "test123").Return(testOrder, nil).Once()
	
	ctx := context.Background()
	result, err := repo.GetOrder(ctx, "test123")
	
	assert.NoError(t, err)
	assert.Equal(t, "test123", result.OrderUID)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "GetOrder", 2)
}

func TestRetryingOrderRepository_PermanentError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	logger, _ := logger.NewLogger(logger.DEV)
	
	retryConfig := &RetryConfig{
		MaxElapsedTime:     1 * time.Second,
		InitialInterval:    10 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:         1.5,
		MaxInterval:        100 * time.Millisecond,
	}
	
	repo := NewRetryingOrderRepository(mockRepo, logger, retryConfig)
	
	testOrder := entities.Order{OrderUID: "test123"}
	
	mockRepo.On("SaveOrder", mock.Anything, testOrder).Return(backoff.Permanent(errors.New("permanent error"))).Once()
	
	ctx := context.Background()
	err := repo.SaveOrder(ctx, testOrder)
	
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "SaveOrder", 1)
}
