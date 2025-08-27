// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka/consumer.go
package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type ConsumerInterface interface {
	Start()
	Shutdown(ctx context.Context) error
}

type Consumer struct {
	reader *kafka.Reader
	svc    application.OrderServiceInterface
	logger logger.LoggerInterface
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewConsumer(brokers []string, topic, groupID string, svc application.OrderServiceInterface, l logger.LoggerInterface) ConsumerInterface {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3,
		MaxBytes:    10e6,
	})

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		reader: r,
		svc:    svc,
		logger: l,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Consumer) Start() {
	c.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("kafka consumer panicked", zap.Any("panic", r))
			}
		}()
		c.consumeLoop()
	}()
}

func (c *Consumer) consumeLoop() {
	defer c.wg.Done()

	for {
		msg, err := c.fetchMessage()
		if err != nil {
			if c.isShuttingDown(err) {
				return
			}
			c.handleFetchError(err)
			continue
		}
		c.processMessage(msg)
	}
}

func (c *Consumer) fetchMessage() (kafka.Message, error) {
	return c.reader.FetchMessage(c.ctx)
}

func (c *Consumer) isShuttingDown(err error) bool {
	return c.ctx.Err() != nil
}

func (c *Consumer) handleFetchError(err error) {
	c.logger.Error("failed to fetch kafka message", zap.Error(err))
	time.Sleep(time.Second)
}

func (c *Consumer) processMessage(msg kafka.Message) {
	order, err := c.decodeOrder(msg.Value)
	if err != nil {
		c.handleDecodeError(msg, err)
		return
	}
	c.saveAndCommit(order, msg)
}

func (c *Consumer) decodeOrder(data []byte) (entities.Order, error) {
	var order entities.Order
	err := json.Unmarshal(data, &order)
	return order, err
}

func (c *Consumer) handleDecodeError(msg kafka.Message, err error) {
	c.logger.Warn("invalid kafka message", zap.ByteString("value", msg.Value), zap.Error(err))
	_ = c.reader.CommitMessages(c.ctx, msg)
}

func (c *Consumer) saveAndCommit(order entities.Order, msg kafka.Message) {
	if _, err := c.svc.SaveOrder(c.ctx, order); err != nil {
		c.logger.Error("failed to save order", zap.String("order_id", order.OrderUID), zap.Error(err))
		return
	}
	if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
		c.logger.Error("failed to commit kafka message", zap.Error(err))
		return
	}
	c.logger.Info("order saved from kafka", zap.String("order_id", order.OrderUID))
}

func (c *Consumer) Shutdown(ctx context.Context) error {
	c.logger.Info("kafka consumer shutting down...")

	c.cancel()

	if err := c.reader.Close(); err != nil {
		c.logger.Error("failed to close kafka reader", zap.Error(err))
		return err
	}

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.Info("kafka consumer stopped gracefully")
		return nil
	case <-ctx.Done():
		c.logger.Warn("kafka consumer shutdown timed out")
		return ctx.Err()
	}
}
