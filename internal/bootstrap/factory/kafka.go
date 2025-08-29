package factory

import (
	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func NewKafkaConsumer(cfg config.KafkaConfig, svc application.OrderServiceInterface, l logger.LoggerInterface) kafka.ConsumerInterface {
	return kafka.NewConsumer(cfg.Brokers, cfg.Topic, cfg.GroupID, svc, l)
}
