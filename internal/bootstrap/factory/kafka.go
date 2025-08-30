// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/kafka.go
package factory

import (
	"time"

	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka"
)

func NewKafkaConsumer(cfg config.KafkaConfig, svc application.OrderServiceInterface, l domainrepo.Logger) domainrepo.EventConsumer {
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
