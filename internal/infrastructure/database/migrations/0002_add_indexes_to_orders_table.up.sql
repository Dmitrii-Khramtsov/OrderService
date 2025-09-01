-- // github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database/migrations/0002_add_indexes_to_orders_table.up.sql
CREATE INDEX IF NOT EXISTS idx_delivery_order_uid ON delivery(order_uid);
CREATE INDEX IF NOT EXISTS idx_payment_order_uid ON payment(order_uid);
CREATE INDEX IF NOT EXISTS idx_orders_order_uid ON orders (order_uid);
CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items (order_uid);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders (date_created);
