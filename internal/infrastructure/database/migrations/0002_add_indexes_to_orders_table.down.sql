-- // github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database/migrations/0002_add_indexes_to_orders_table.down.sql
DROP INDEX IF EXISTS idx_delivery_order_uid;
DROP INDEX IF EXISTS idx_payment_order_uid;
DROP INDEX IF EXISTS idx_orders_order_uid;
DROP INDEX IF EXISTS idx_items_order_uid;
DROP INDEX IF EXISTS idx_orders_created_at;
