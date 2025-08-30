// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database/postgres_repository.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
)

type PostgresOrderRepository struct {
	db     *sqlx.DB
	logger domainrepo.Logger
}

func NewPostgresOrderRepository(db *sqlx.DB, logger domainrepo.Logger) (*PostgresOrderRepository, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	return &PostgresOrderRepository{db: db, logger: logger}, nil
}

func (r *PostgresOrderRepository) SaveOrder(ctx context.Context, order entities.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", domainrepo.ErrTransactionFailed, err)
	}
	defer tx.Rollback()

	if err := r.saveOrder(ctx, tx, order); err != nil {
		return err
	}

	if err := r.saveDelivery(ctx, tx, order); err != nil {
		return err
	}

	if err := r.savePayment(ctx, tx, order); err != nil {
		return err
	}

	if err := r.saveItems(ctx, tx, order); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresOrderRepository) saveOrder(ctx context.Context, tx *sqlx.Tx, order entities.Order) error {
	query := `
		INSERT INTO orders (
				order_uid, track_number, entry, locale, internal_signature,
				customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES (
				:order_uid, :track_number, :entry, :locale, :internal_signature,
				:customer_id, :delivery_service, :shardkey, :sm_id, :date_created, :oof_shard
		) ON CONFLICT (order_uid) DO UPDATE SET
				track_number = EXCLUDED.track_number,
				entry = EXCLUDED.entry,
				locale = EXCLUDED.locale,
				internal_signature = EXCLUDED.internal_signature,
				customer_id = EXCLUDED.customer_id,
				delivery_service = EXCLUDED.delivery_service,
				shardkey = EXCLUDED.shardkey,
				sm_id = EXCLUDED.sm_id,
				date_created = EXCLUDED.date_created,
				oof_shard = EXCLUDED.oof_shard
    `

	_, err := tx.NamedExecContext(ctx, query, order)
	if err != nil {
		r.logger.Error("failed to save order", "error", err, "order_uid", order.OrderUID)
		return fmt.Errorf("%w: %v", domainrepo.ErrOrderSaveFailed, err)
	}

	return nil
}

func (r *PostgresOrderRepository) saveDelivery(ctx context.Context, tx *sqlx.Tx, order entities.Order) error {
	query := `
		INSERT INTO delivery (
				order_uid, name, phone, zip, city, address, region, email
		) VALUES (
				:order_uid, :name, :phone, :zip, :city, :address, :region, :email
		) ON CONFLICT (order_uid) DO UPDATE SET
				name = EXCLUDED.name,
				phone = EXCLUDED.phone,
				zip = EXCLUDED.zip,
				city = EXCLUDED.city,
				address = EXCLUDED.address,
				region = EXCLUDED.region,
				email = EXCLUDED.email
    `

	deliveryMap := map[string]interface{}{
		"order_uid": order.OrderUID,
		"name":      order.Delivery.Name,
		"phone":     order.Delivery.Phone,
		"zip":       order.Delivery.Zip,
		"city":      order.Delivery.City,
		"address":   order.Delivery.Address,
		"region":    order.Delivery.Region,
		"email":     order.Delivery.Email,
	}

	_, err := tx.NamedExecContext(ctx, query, deliveryMap)
	if err != nil {
		r.logger.Error("failed to save delivery", "error", err, "order_uid", order.OrderUID)
		return fmt.Errorf("%w: %v", domainrepo.ErrOrderSaveFailed, err)
	}

	return nil
}

func (r *PostgresOrderRepository) savePayment(ctx context.Context, tx *sqlx.Tx, order entities.Order) error {
	query := `
		INSERT INTO payment (
				order_uid, transaction, request_id, currency, provider,
				amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES (
				:order_uid, :transaction, :request_id, :currency, :provider,
				:amount, :payment_dt, :bank, :delivery_cost, :goods_total, :custom_fee
		) ON CONFLICT (order_uid) DO UPDATE SET
				transaction = EXCLUDED.transaction,
				request_id = EXCLUDED.request_id,
				currency = EXCLUDED.currency,
				provider = EXCLUDED.provider,
				amount = EXCLUDED.amount,
				payment_dt = EXCLUDED.payment_dt,
				bank = EXCLUDED.bank,
				delivery_cost = EXCLUDED.delivery_cost,
				goods_total = EXCLUDED.goods_total,
				custom_fee = EXCLUDED.custom_fee
    `

	paymentMap := map[string]interface{}{
		"order_uid":     order.OrderUID,
		"transaction":   order.Payment.Transaction,
		"request_id":    order.Payment.RequestID,
		"currency":      order.Payment.Currency,
		"provider":      order.Payment.Provider,
		"amount":        order.Payment.Amount,
		"payment_dt":    order.Payment.PaymentDT,
		"bank":          order.Payment.Bank,
		"delivery_cost": order.Payment.DeliveryCost,
		"goods_total":   order.Payment.GoodsTotal,
		"custom_fee":    order.Payment.CustomFee,
	}

	_, err := tx.NamedExecContext(ctx, query, paymentMap)
	if err != nil {
		r.logger.Error("failed to save payment", "error", err, "order_uid", order.OrderUID)
		return fmt.Errorf("%w: %v", domainrepo.ErrOrderSaveFailed, err)
	}

	return nil
}

func (r *PostgresOrderRepository) saveItems(ctx context.Context, tx *sqlx.Tx, order entities.Order) error {
	deleteQuery := "DELETE FROM items WHERE order_uid = $1"
	_, err := tx.ExecContext(ctx, deleteQuery, order.OrderUID)
	if err != nil {
		r.logger.Error("failed to delete existing items", "error", err, "order_uid", order.OrderUID)
		return fmt.Errorf("%w: %v", domainrepo.ErrOrderSaveFailed, err)
	}

	query := `
		INSERT INTO items (
				chrt_id, order_uid, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status
		) VALUES (
				:chrt_id, :order_uid, :track_number, :price, :rid, :name,
				:sale, :size, :total_price, :nm_id, :brand, :status
		)
    `

	for _, item := range order.Items {
		itemMap := map[string]interface{}{
			"chrt_id":      item.ChrtID,
			"order_uid":    order.OrderUID,
			"track_number": item.TrackNumber,
			"price":        item.Price,
			"rid":          item.RID,
			"name":         item.Name,
			"sale":         item.Sale,
			"size":         item.Size,
			"total_price":  item.TotalPrice,
			"nm_id":        item.NmID,
			"brand":        item.Brand,
			"status":       item.Status,
		}

		_, err := tx.NamedExecContext(ctx, query, itemMap)
		if err != nil {
			r.logger.Error("failed to save item", "error", err, "order_uid", order.OrderUID)
			return fmt.Errorf("%w: %v", domainrepo.ErrOrderSaveFailed, err)
		}
	}

	return nil
}

func (r *PostgresOrderRepository) GetOrder(ctx context.Context, id string) (entities.Order, error) {
	query := `
		SELECT 
				o.*,
				d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
				p.transaction, p.request_id, p.currency, p.provider, p.amount,
				p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
				i.chrt_id, i.track_number, i.price, i.rid, i.name as item_name,
				i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status
		FROM orders o
		LEFT JOIN delivery d ON o.order_uid = d.order_uid
		LEFT JOIN payment p ON o.order_uid = p.order_uid
		LEFT JOIN items i ON o.order_uid = i.order_uid
		WHERE o.order_uid = $1
    `

	rows, err := r.db.QueryxContext(ctx, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Order{}, domainrepo.ErrOrderNotFound
		}
		r.logger.Error("failed to get order", "error", err, "order_uid", id)
		return entities.Order{}, fmt.Errorf("%w: %v", domainrepo.ErrQueryFailed, err)
	}
	defer rows.Close()

	var order entities.Order
	var items []entities.Item

	for rows.Next() {
		var item entities.Item
		var delivery entities.Delivery
		var payment entities.Payment

		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSig, &order.CustomerID, &order.DeliveryService,
			&order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
			&delivery.Address, &delivery.Region, &delivery.Email,
			&payment.Transaction, &payment.RequestID, &payment.Currency,
			&payment.Provider, &payment.Amount, &payment.PaymentDT, &payment.Bank,
			&payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)

		if err != nil {
			r.logger.Error("failed to scan order", "error", err, "order_uid", id)
			return entities.Order{}, fmt.Errorf("%w: %v", domainrepo.ErrQueryFailed, err)
		}

		order.Delivery = delivery
		order.Payment = payment
		items = append(items, item)
	}

	if order.OrderUID == "" {
		return entities.Order{}, domainrepo.ErrOrderNotFound
	}

	order.Items = items
	return order, nil
}

func (r *PostgresOrderRepository) GetAllOrders(ctx context.Context, limit, offset int) ([]entities.Order, error) {
	// основной запрос для получения заказов с доставкой и оплатой
	mainQuery := `
		SELECT 
				o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
				o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
				d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
				p.transaction, p.request_id, p.currency, p.provider, p.amount,
				p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
		FROM orders o
		LEFT JOIN delivery d ON o.order_uid = d.order_uid
		LEFT JOIN payment p ON o.order_uid = p.order_uid
		ORDER BY o.order_uid
		LIMIT $1 OFFSET $2
  `

	rows, err := r.db.QueryContext(ctx, mainQuery, limit, offset)
	if err != nil {
		r.logger.Error("failed to get all orders", "error", err)
		return nil, fmt.Errorf("%w: %v", domainrepo.ErrQueryFailed, err)
	}
	defer rows.Close()

	var orders []entities.Order
	var orderUIDs []string
	ordersMap := make(map[string]*entities.Order)

	for rows.Next() {
		var o entities.Order
		var d entities.Delivery
		var p entities.Payment

		err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSig,
			&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SMID, &o.DateCreated, &o.OOFShard,
			&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
			&p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount,
			&p.PaymentDT, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
		)
		if err != nil {
			r.logger.Error("failed to scan order", "error", err)
			continue
		}

		o.Delivery = d
		o.Payment = p
		orders = append(orders, o)
		orderUIDs = append(orderUIDs, o.OrderUID)
		ordersMap[o.OrderUID] = &orders[len(orders)-1]
	}

	if len(orderUIDs) == 0 {
		return orders, nil
	}

	// второй запрос для получения всех items для найденных заказов
	itemsQuery := `
		SELECT 
				chrt_id, order_uid, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = ANY($1)
		ORDER BY order_uid, chrt_id
  `

	itemRows, err := r.db.QueryContext(ctx, itemsQuery, pq.Array(orderUIDs))
	if err != nil {
		r.logger.Error("failed to get items for orders", "error", err)
		return nil, fmt.Errorf("%w: %v", domainrepo.ErrQueryFailed, err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item entities.Item
		var orderUID string

		err := itemRows.Scan(
			&item.ChrtID, &orderUID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			r.logger.Error("failed to scan item", "error", err)
			continue
		}

		if order, exists := ordersMap[orderUID]; exists {
			order.Items = append(order.Items, item)
		}
	}

	return orders, nil
}

func (r *PostgresOrderRepository) GetOrdersCount(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM orders"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		r.logger.Error("failed to get orders count", "error", err)
		return 0, fmt.Errorf("%w: %v", domainrepo.ErrQueryFailed, err)
	}
	return count, nil
}

func (r *PostgresOrderRepository) DeleteOrder(ctx context.Context, id string) error {
	query := "DELETE FROM orders WHERE order_uid = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete order", "error", err, "order_uid", id)
		return fmt.Errorf("%w: %v", domainrepo.ErrOrderDeleteFailed, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", domainrepo.ErrOrderDeleteFailed, err)
	}

	if rowsAffected == 0 {
		return domainrepo.ErrOrderNotFound
	}

	return nil
}

func (r *PostgresOrderRepository) ClearOrders(ctx context.Context) error {
	query := "DELETE FROM orders"
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		r.logger.Error("failed to clear orders", "error", err)
		return fmt.Errorf("%w: %v", domainrepo.ErrOrderClearFailed, err)
	}
	return nil
}

func (r *PostgresOrderRepository) Shutdown(ctx context.Context) error {
	return r.db.Close()
}
