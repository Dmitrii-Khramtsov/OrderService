// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory/kafka.go
package factory

import (
	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka"
)

func NewKafkaConsumer(cfg config.KafkaConfig, svc application.OrderServiceInterface, l domainrepo.Logger) domainrepo.EventConsumer {
	retryConfig := &kafka.RetryConfig{
		InitialInterval:     cfg.Retry.InitialInterval,
		Multiplier:          cfg.Retry.Multiplier,
		MaxInterval:         cfg.Retry.MaxInterval,
		MaxElapsedTime:      cfg.Retry.MaxElapsedTime,
		RandomizationFactor: cfg.Retry.RandomizationFactor,
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
		cfg.ProcessingTime,
		cfg.MinBytes,
		cfg.MaxBytes,
		cfg.MaxWait,
		cfg.CommitInterval,
		cfg.BatchTimeout,
		cfg.BatchSize,
	)
}
