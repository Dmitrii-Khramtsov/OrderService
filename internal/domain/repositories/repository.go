// github.com/Dmitrii-Khramtsov/orderservice/internal/repositories/entities
package repositories

import "database/sql"

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository {
		db: db,
	}
}
