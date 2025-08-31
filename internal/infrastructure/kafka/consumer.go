// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka/consumer.go
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/segmentio/kafka-go"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

type Consumer struct {
	reader         *kafka.Reader
	dlqWriter      *kafka.Writer
	svc            application.OrderServiceInterface
	logger         domainrepo.Logger
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	retryConfig    *RetryConfig
	maxRetries     int
	processingTime time.Duration
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
	processingTime time.Duration,
) domainrepo.EventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		StartOffset:    kafka.FirstOffset,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		CommitInterval: 1 * time.Second,
	})

	dlqWriter := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  dlqTopic,
		Balancer:               &kafka.LeastBytes{},
		BatchTimeout:           100 * time.Millisecond,
		BatchSize:              1,
		AllowAutoTopicCreation: true,
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		reader:         reader,
		dlqWriter:      dlqWriter,
		svc:            svc,
		logger:         l,
		ctx:            ctx,
		cancel:         cancel,
		retryConfig:    retryConfig,
		maxRetries:     maxRetries,
		processingTime: processingTime,
	}
}

func (c *Consumer) Start() {
	c.logger.Info("starting Kafka consumer",
		"topic", c.reader.Config().Topic,
		"group_id", c.reader.Config().GroupID,
	)

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
			c.logger.Info("Kafka consumer loop stopped")
			return
		default:
			c.processMessage()
		}
	}
}

func (c *Consumer) processMessage() {
	msg, err := c.reader.FetchMessage(c.ctx)
	if err != nil {
		if c.ctx.Err() != nil {
			return
		}
		c.logger.Error("failed to fetch Kafka message", "error", err)
		time.Sleep(1 * time.Second)
		return
	}

	c.logger.Debug("received Kafka message",
		"key", string(msg.Key),
		"topic", msg.Topic,
		"partition", msg.Partition,
		"offset", msg.Offset,
	)

	startTime := time.Now()
	err = c.processWithRetry(msg)
	processingTime := time.Since(startTime)

	if err != nil {
		c.handleProcessingError(msg, err, processingTime)
		return
	}

	if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
		c.logger.Error("failed to commit Kafka message",
			"key", string(msg.Key),
			"error", err,
			"processing_time", processingTime,
		)
		return
	}

	c.logger.Info("successfully processed Kafka message",
		"key", string(msg.Key),
		"processing_time", processingTime,
	)
}

func (c *Consumer) processWithRetry(msg kafka.Message) error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = c.retryConfig.InitialInterval
	expBackoff.Multiplier = c.retryConfig.Multiplier
	expBackoff.MaxInterval = c.retryConfig.MaxInterval
	expBackoff.MaxElapsedTime = c.retryConfig.MaxElapsedTime
	expBackoff.RandomizationFactor = c.retryConfig.RandomizationFactor

	ctx, cancel := context.WithTimeout(c.ctx, c.processingTime)
	defer cancel()

	var lastErr error
	retries := 0

	operation := func() error {
		retries++
		order, err := c.decodeOrder(msg.Value)
		if err != nil {
			return backoff.Permanent(err)
		}

		_, err = c.svc.SaveOrder(ctx, order)
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

func (c *Consumer) handleProcessingError(msg kafka.Message, err error, processingTime time.Duration) {
	c.logger.Error("failed to process message after retries, sending to DLQ",
		"key", string(msg.Key),
		"topic", msg.Topic,
		"partition", msg.Partition,
		"offset", msg.Offset,
		"error", err,
		"processing_time", processingTime,
	)

	if err := c.sendToDLQ(msg); err != nil {
		c.logger.Error("failed to send message to DLQ",
			"key", string(msg.Key),
			"error", err,
		)
	}
}

func (c *Consumer) decodeOrder(data []byte) (entities.Order, error) {
	var order entities.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return entities.Order{}, fmt.Errorf("%w: %v", ErrKafkaMessageDecode, err)
	}

	if err := order.Validate(); err != nil {
		return entities.Order{}, fmt.Errorf("%w: %v", domain.ErrInvalidOrder, err)
	}

	return order, nil
}

func (c *Consumer) sendToDLQ(msg kafka.Message) error {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	dlqMsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Time:  msg.Time,
		Headers: append(msg.Headers, kafka.Header{
			Key:   "original_topic",
			Value: []byte(msg.Topic),
		}),
	}

	if err := c.dlqWriter.WriteMessages(ctx, dlqMsg); err != nil {
		return fmt.Errorf("%w: %v", ErrKafkaMessageSend, err)
	}

	c.logger.Info("message sent to DLQ",
		"key", string(msg.Key),
		"original_topic", msg.Topic,
	)
	return nil
}

func (c *Consumer) Shutdown(ctx context.Context) error {
	c.logger.Info("Kafka consumer shutting down...")
	c.cancel()

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	shutdownErr := make(chan error, 2)

	go func() {
		if err := c.reader.Close(); err != nil {
			c.logger.Error("failed to close Kafka reader", "error", err)
			shutdownErr <- fmt.Errorf("%w: %v", ErrKafkaConnectionFailed, err)
		} else {
			shutdownErr <- nil
		}
	}()

	go func() {
		if err := c.dlqWriter.Close(); err != nil {
			c.logger.Error("failed to close DLQ writer", "error", err)
			shutdownErr <- fmt.Errorf("%w: %v", ErrKafkaConnectionFailed, err)
		} else {
			shutdownErr <- nil
		}
	}()

	// Wait for both close operations
	var firstErr error
	for i := 0; i < 2; i++ {
		select {
		case err := <-shutdownErr:
			if err != nil && firstErr == nil {
				firstErr = err
			}
		case <-shutdownCtx.Done():
			c.logger.Warn("Kafka consumer shutdown timed out")
			return shutdownCtx.Err()
		}
	}

	// Wait for consumer goroutine to finish
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.Info("Kafka consumer stopped gracefully")
		return firstErr
	case <-shutdownCtx.Done():
		c.logger.Warn("Kafka consumer shutdown timed out while waiting for goroutines")
		return shutdownCtx.Err()
	}
}
