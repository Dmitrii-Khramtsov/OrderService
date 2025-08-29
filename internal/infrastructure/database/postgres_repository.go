// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database/postgres_repository.go
package repository

import (
	"context"
	"database/sql"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PostgresOrderRepository struct {
	db     *sqlx.DB
	logger logger.LoggerInterface
}

func NewPostgresOrderRepository(db *sqlx.DB, l logger.LoggerInterface) (OrderRepository, error) {
	return &PostgresOrderRepository{
		db:     db,
		logger: l,
	}, nil
}

func (r *PostgresOrderRepository) SaveOrder(ctx context.Context, order entities.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("begin transaction failed", zap.Error(err))
		return ErrOrderSaveFailed
	}
	defer tx.Rollback()

	if err := r.saveOrderRow(ctx, tx, order); err != nil {
		return ErrOrderSaveFailed
	}
	if err := r.saveDeliveryRow(ctx, tx, order); err != nil {
		return ErrOrderSaveFailed
	}
	if err := r.savePaymentRow(ctx, tx, order); err != nil {
		return ErrOrderSaveFailed
	}
	if err := r.saveItemsRows(ctx, tx, order); err != nil {
		return ErrOrderSaveFailed
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("commit transaction failed", zap.Error(err))
		return ErrOrderSaveFailed
	}
	return nil
}

func (r *PostgresOrderRepository) GetOrder(ctx context.Context, id string) (entities.Order, error) {
	order, err := r.getOrderRow(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.Order{}, ErrOrderNotFound
		}
		return entities.Order{}, ErrQueryFailed
	}

	order.Delivery, _ = r.getDeliveryRow(ctx, id)
	order.Payment, _ = r.getPaymentRow(ctx, id)
	order.Items, _ = r.getItemsRows(ctx, id)

	return order, nil
}

func (r *PostgresOrderRepository) GetAllOrders(ctx context.Context) ([]entities.Order, error) {
	var orders []entities.Order
	if err := r.db.SelectContext(ctx, &orders, "SELECT * FROM orders"); err != nil {
		r.logger.Error("failed to get all orders", zap.Error(err))
		return nil, ErrQueryFailed
	}

	for i := range orders {
		orderID := orders[i].OrderUID
		orders[i].Delivery, _ = r.getDeliveryRow(ctx, orderID)
		orders[i].Payment, _ = r.getPaymentRow(ctx, orderID)
		orders[i].Items, _ = r.getItemsRows(ctx, orderID)
	}

	return orders, nil
}

func (r *PostgresOrderRepository) DeleteOrder(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM orders WHERE order_uid=$1", id)
	if err != nil {
		r.logger.Error("failed to delete order", zap.String("order_uid", id), zap.Error(err))
		return ErrOrderDeleteFailed
	}
	return nil
}

func (r *PostgresOrderRepository) ClearOrders(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "TRUNCATE TABLE items, payment, delivery, orders CASCADE")
	if err != nil {
		r.logger.Error("failed to clear orders", zap.Error(err))
		return ErrOrderClearFailed
	}
	return nil
}

func (r *PostgresOrderRepository) Shutdown(ctx context.Context) error {
	return r.db.Close()
}

func (r *PostgresOrderRepository) saveOrderRow(ctx context.Context, tx *sqlx.Tx, order entities.Order) error {
	_, err := tx.ExecContext(ctx, `
	INSERT INTO orders(order_uid, track_number, entry, locale, internal_signature, customer_id,
		delivery_service, shardkey, sm_id, date_created, oof_shard)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	ON CONFLICT(order_uid) DO UPDATE SET
	track_number=EXCLUDED.track_number,
	entry=EXCLUDED.entry,
	locale=EXCLUDED.locale,
	internal_signature=EXCLUDED.internal_signature,
	customer_id=EXCLUDED.customer_id,
	delivery_service=EXCLUDED.delivery_service,
	shardkey=EXCLUDED.shardkey,
	sm_id=EXCLUDED.sm_id,
	date_created=EXCLUDED.date_created,
	oof_shard=EXCLUDED.oof_shard
	`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSig, order.CustomerID, order.DeliveryService, order.ShardKey,
		order.SMID, order.DateCreated, order.OOFShard)
	return err
}

func (r *PostgresOrderRepository) saveDeliveryRow(ctx context.Context, tx *sqlx.Tx, order entities.Order) error {
	_, err := tx.ExecContext(ctx, `
	INSERT INTO delivery(order_uid, name, phone, zip, city, address, region, email)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8)
	ON CONFLICT(order_uid) DO UPDATE SET
	name=EXCLUDED.name,
	phone=EXCLUDED.phone,
	zip=EXCLUDED.zip,
	city=EXCLUDED.city,
	address=EXCLUDED.address,
	region=EXCLUDED.region,
	email=EXCLUDED.email
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	return err
}

func (r *PostgresOrderRepository) savePaymentRow(ctx context.Context, tx *sqlx.Tx, order entities.Order) error {
	_, err := tx.ExecContext(ctx, `
	INSERT INTO payment(order_uid, transaction, request_id, currency, provider, amount, payment_dt,
		bank, delivery_cost, goods_total, custom_fee)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	ON CONFLICT(order_uid) DO UPDATE SET
	transaction=EXCLUDED.transaction,
	request_id=EXCLUDED.request_id,
	currency=EXCLUDED.currency,
	provider=EXCLUDED.provider,
	amount=EXCLUDED.amount,
	payment_dt=EXCLUDED.payment_dt,
	bank=EXCLUDED.bank,
	delivery_cost=EXCLUDED.delivery_cost,
	goods_total=EXCLUDED.goods_total,
	custom_fee=EXCLUDED.custom_fee
	`, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT,
		order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	return err
}

func (r *PostgresOrderRepository) saveItemsRows(ctx context.Context, tx *sqlx.Tx, order entities.Order) error {
	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO items(chrt_id, order_uid, track_number, price, rid, name, sale, size,
			total_price, nm_id, brand, status)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT(chrt_id, order_uid) DO UPDATE SET
			track_number=EXCLUDED.track_number,
			price=EXCLUDED.price,
			rid=EXCLUDED.rid,
			name=EXCLUDED.name,
			sale=EXCLUDED.sale,
			size=EXCLUDED.size,
			total_price=EXCLUDED.total_price,
			nm_id=EXCLUDED.nm_id,
			brand=EXCLUDED.brand,
			status=EXCLUDED.status
		`, item.ChrtID, order.OrderUID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresOrderRepository) getOrderRow(ctx context.Context, id string) (entities.Order, error) {
	var order entities.Order
	err := r.db.GetContext(ctx, &order, "SELECT * FROM orders WHERE order_uid=$1", id)
	if err != nil {
		r.logger.Error("failed to get order", zap.String("order_uid", id), zap.Error(err))
		return entities.Order{}, err
	}
	return order, nil
}

func (r *PostgresOrderRepository) getDeliveryRow(ctx context.Context, id string) (entities.Delivery, error) {
	var delivery entities.Delivery
	err := r.db.GetContext(ctx, &delivery, "SELECT * FROM delivery WHERE order_uid=$1", id)
	if err != nil && err != sql.ErrNoRows {
		r.logger.Error("failed to get delivery", zap.String("order_uid", id), zap.Error(err))
		return entities.Delivery{}, err
	}
	return delivery, nil
}

func (r *PostgresOrderRepository) getPaymentRow(ctx context.Context, id string) (entities.Payment, error) {
	var payment entities.Payment
	err := r.db.GetContext(ctx, &payment, "SELECT * FROM payment WHERE order_uid=$1", id)
	if err != nil && err != sql.ErrNoRows {
		r.logger.Error("failed to get payment", zap.String("order_uid", id), zap.Error(err))
		return entities.Payment{}, err
	}
	return payment, nil
}

func (r *PostgresOrderRepository) getItemsRows(ctx context.Context, id string) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.SelectContext(ctx, &items, "SELECT * FROM items WHERE order_uid=$1", id)
	if err != nil && err != sql.ErrNoRows {
		r.logger.Error("failed to get items", zap.String("order_uid", id), zap.Error(err))
		return nil, err
	}
	return items, nil
}
