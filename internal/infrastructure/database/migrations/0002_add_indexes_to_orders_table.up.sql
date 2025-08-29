-- // github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database/migrations/0002_add_indexes_to_orders_table.down.down.sql
CREATE INDEX idx_delivery_order_uid ON delivery(order_uid);
CREATE INDEX idx_payment_order_uid ON payment(order_uid);
