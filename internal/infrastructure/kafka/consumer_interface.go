// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka/consumer_interface.go
package kafka

import "context"

type ConsumerInterface interface {
	Start()
	Shutdown(ctx context.Context) error
}
