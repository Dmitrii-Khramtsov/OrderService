// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository/event_consumer.go
package repository

import "context"

type EventConsumer interface {
	Start()
	Shutdown(ctx context.Context) error
}
