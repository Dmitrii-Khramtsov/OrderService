// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka/consumer.go
package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/segmentio/kafka-go"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

type Consumer struct {
	reader      *kafka.Reader
	dlqWriter   *kafka.Writer
	svc         application.OrderServiceInterface
	logger      domainrepo.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	retryConfig *RetryConfig
	maxRetries  int
}

type RetryConfig struct {
	InitialInterval     time.Duration
	Multiplier          float64
	MaxInterval         time.Duration
	MaxElapsedTime      time.Duration
	RandomizationFactor float64
}

func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	dlqTopic string,
	svc application.OrderServiceInterface,
	l domainrepo.Logger,
	retryConfig *RetryConfig,
	maxRetries int,
) domainrepo.EventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3,
		MaxBytes:    10e6,
	})

	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    dlqTopic,
		Balancer: &kafka.LeastBytes{},
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		reader:      reader,
		dlqWriter:   dlqWriter,
		svc:         svc,
		logger:      l,
		ctx:         ctx,
		cancel:      cancel,
		retryConfig: retryConfig,
		maxRetries:  maxRetries,
	}
}

func (c *Consumer) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consumeLoop()
	}()
}

func (c *Consumer) consumeLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.reader.FetchMessage(c.ctx)
			if err != nil {
				if c.ctx.Err() != nil {
					return
				}
				c.logger.Error("failed to fetch message", "error", err)
				time.Sleep(time.Second)
				continue
			}

			if err := c.processWithRetry(msg); err != nil {
				c.logger.Error("failed to process message after retries, sending to DLQ",
					"topic", c.reader.Config().Topic,
					"key", string(msg.Key),
					"error", err,
				)

				if err := c.sendToDLQ(msg); err != nil {
					c.logger.Error("failed to send message to DLQ",
						"key", string(msg.Key),
						"error", err,
					)
				}
			}

			if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
				c.logger.Error("failed to commit message",
					"key", string(msg.Key),
					"error", err,
				)
			}
		}
	}
}

func (c *Consumer) processWithRetry(msg kafka.Message) error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = c.retryConfig.InitialInterval
	expBackoff.Multiplier = c.retryConfig.Multiplier
	expBackoff.MaxInterval = c.retryConfig.MaxInterval
	expBackoff.MaxElapsedTime = c.retryConfig.MaxElapsedTime
	expBackoff.RandomizationFactor = c.retryConfig.RandomizationFactor

	var lastErr error
	retries := 0

	operation := func() error {
		retries++
		order, err := c.decodeOrder(msg.Value)
		if err != nil {
			return backoff.Permanent(err)
		}

		_, err = c.svc.SaveOrder(c.ctx, order)
		if err != nil {
			lastErr = err
			c.logger.Warn("failed to process message, retrying",
				"order_uid", order.OrderUID,
				"attempt", retries,
				"error", err,
			)
		}
		return err
	}

	err := backoff.Retry(operation, expBackoff)
	if err != nil && retries >= c.maxRetries {
		return lastErr
	}
	return err
}

func (c *Consumer) decodeOrder(data []byte) (entities.Order, error) {
	var order entities.Order
	err := json.Unmarshal(data, &order)
	return order, err
}

func (c *Consumer) sendToDLQ(msg kafka.Message) error {
	return c.dlqWriter.WriteMessages(c.ctx, kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Time:  msg.Time,
	})
}

func (c *Consumer) Shutdown(ctx context.Context) error {
	c.logger.Info("kafka consumer shutting down...")
	c.cancel()

	if err := c.reader.Close(); err != nil {
		c.logger.Error("failed to close kafka reader", "error", err)
		return err
	}

	if err := c.dlqWriter.Close(); err != nil {
		c.logger.Error("failed to close DLQ writer", "error", err)
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
