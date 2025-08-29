// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/kafka.go
package factory

import (
	"time"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func NewKafkaConsumer(cfg config.KafkaConfig, svc application.OrderServiceInterface, l logger.LoggerInterface) kafka.ConsumerInterface {
	retryConfig := &kafka.RetryConfig{
		InitialInterval:    time.Second,
		Multiplier:         2,
		MaxInterval:        30 * time.Second,
		MaxElapsedTime:     5 * time.Minute,
		RandomizationFactor: 0.5,
	}

	return kafka.NewConsumer(
		cfg.Brokers,
		cfg.Topic,
		cfg.GroupID,
		cfg.DLQTopic,
		svc,
		l,
		retryConfig,
		cfg.MaxRetries,
	)
}
