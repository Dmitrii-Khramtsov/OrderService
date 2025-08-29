// internal/infrastructure/database/postgres_repository_integration_test.go
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	ctx := context.Background()
	
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)
	
	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)
	
	dsn := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"
	
	time.Sleep(3 * time.Second)
	
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	
	_, err = db.Exec(`
		CREATE TABLE orders (
			order_uid TEXT PRIMARY KEY NOT NULL,
			track_number TEXT NOT NULL,
			entry TEXT NOT NULL,
			locale TEXT NOT NULL,
			internal_signature TEXT,
			customer_id TEXT NOT NULL,
			delivery_service TEXT NOT NULL,
			shardkey TEXT,
			sm_id INTEGER NOT NULL,
			date_created TIMESTAMP NOT NULL,
			oof_shard TEXT
		);
		
		CREATE TABLE delivery (
			order_uid TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			phone TEXT NOT NULL,
			zip TEXT NOT NULL,
			city TEXT NOT NULL,
			address TEXT NOT NULL,
			region TEXT NOT NULL,
			email TEXT
		);
		
		CREATE TABLE payment (
			order_uid TEXT PRIMARY KEY NOT NULL,
			transaction TEXT NOT NULL,
			request_id TEXT,
			currency TEXT NOT NULL,
			provider TEXT NOT NULL,
			amount INTEGER NOT NULL,
			payment_dt BIGINT NOT NULL,
			bank TEXT,
			delivery_cost INTEGER NOT NULL,
			goods_total INTEGER NOT NULL,
			custom_fee INTEGER NOT NULL
		);
		
		CREATE TABLE items (
			chrt_id INTEGER NOT NULL,
			order_uid TEXT NOT NULL,
			track_number TEXT NOT NULL,
			price INTEGER NOT NULL,
			rid TEXT NOT NULL,
			name TEXT NOT NULL,
			sale INTEGER NOT NULL,
			size TEXT NOT NULL,
			total_price INTEGER NOT NULL,
			nm_id INTEGER NOT NULL,
			brand TEXT NOT NULL,
			status INTEGER NOT NULL,
			PRIMARY KEY (chrt_id, order_uid)
		);
	`)
	require.NoError(t, err)
	
	t.Cleanup(func() {
		db.Close()
		postgresContainer.Terminate(ctx)
	})
	
	return db
}

func TestPostgresOrderRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	db := setupTestDB(t)
	logger, _ := logger.NewLogger(logger.DEV)
	
	repo, err := NewPostgresOrderRepository(db, logger)
	require.NoError(t, err)
	
	t.Run("Save and Get Order", func(t *testing.T) {
		ctx := context.Background()
		
		order := entities.Order{
			OrderUID:        "test-order-1",
			TrackNumber:     "TEST123",
			Entry:           "WBIL",
			Locale:          "en",
			InternalSig:     "",
			CustomerID:      "test-customer",
			DeliveryService: "meest",
			ShardKey:        "9",
			SMID:            99,
			DateCreated:     "2021-11-26T06:22:19Z",
			OOFShard:        "1",
			Delivery: entities.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: entities.Payment{
				Transaction:  "test-transaction-1",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1817,
				PaymentDT:    1637907727,
				Bank:         "alpha",
				DeliveryCost: 1500,
				GoodsTotal:   317,
				CustomFee:    0,
			},
			Items: []entities.Item{
				{
					ChrtID:      9934930,
					TrackNumber: "TEST123",
					Price:       453,
					RID:         "ab4219087a764ae0btest",
					Name:        "Mascaras",
					Sale:        30,
					Size:        "0",
					TotalPrice:  317,
					NmID:        2389212,
					Brand:       "Vivienne Sabo",
					Status:      202,
				},
			},
		}
		
		err := repo.SaveOrder(ctx, order)
		assert.NoError(t, err)
		
		retrieved, err := repo.GetOrder(ctx, "test-order-1")
		assert.NoError(t, err)
		assert.Equal(t, order.OrderUID, retrieved.OrderUID)
		assert.Equal(t, order.TrackNumber, retrieved.TrackNumber)
		assert.Len(t, retrieved.Items, 1)
		assert.Equal(t, order.Items[0].ChrtID, retrieved.Items[0].ChrtID)
		
		orders, err := repo.GetAllOrders(ctx, 100, 0)
		assert.NoError(t, err)
		assert.Len(t, orders, 1)
		assert.Equal(t, order.OrderUID, orders[0].OrderUID)
	})
	
	t.Run("Get Non-Existent Order", func(t *testing.T) {
		ctx := context.Background()
		
		_, err := repo.GetOrder(ctx, "non-existent")
		assert.ErrorIs(t, err, ErrOrderNotFound)
	})
	
	t.Run("Update Order", func(t *testing.T) {
		ctx := context.Background()
		
		order := entities.Order{
			OrderUID:        "test-order-2",
			TrackNumber:     "TEST456",
			Entry:           "WBIL",
			Locale:          "en",
			InternalSig:     "",
			CustomerID:      "test-customer",
			DeliveryService: "meest",
			ShardKey:        "9",
			SMID:            99,
			DateCreated:     "2021-11-26T06:22:19Z",
			OOFShard:        "1",
			Delivery: entities.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: entities.Payment{
				Transaction:  "test-transaction-2",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1817,
				PaymentDT:    1637907727,
				Bank:         "alpha",
				DeliveryCost: 1500,
				GoodsTotal:   317,
				CustomFee:    0,
			},
			Items: []entities.Item{
				{
					ChrtID:      9934931,
					TrackNumber: "TEST456",
					Price:       453,
					RID:         "ab4219087a764ae0btest",
					Name:        "Mascaras",
					Sale:        30,
					Size:        "0",
					TotalPrice:  317,
					NmID:        2389212,
					Brand:       "Vivienne Sabo",
					Status:      202,
				},
			},
		}
		
		err := repo.SaveOrder(ctx, order)
		assert.NoError(t, err)
		
		order.Delivery.Name = "Updated Name"
		order.Payment.Amount = 2000
		order.Items[0].Price = 500
		
		err = repo.SaveOrder(ctx, order)
		assert.NoError(t, err)
		
		retrieved, err := repo.GetOrder(ctx, "test-order-2")
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Delivery.Name)
		assert.Equal(t, 2000, retrieved.Payment.Amount)
		assert.Equal(t, 500, retrieved.Items[0].Price)
	})
	
	t.Run("Delete Order", func(t *testing.T) {
		ctx := context.Background()
		
		order := entities.Order{
			OrderUID:        "test-order-3",
			TrackNumber:     "TEST789",
			Entry:           "WBIL",
			Locale:          "en",
			InternalSig:     "",
			CustomerID:      "test-customer",
			DeliveryService: "meest",
			ShardKey:        "9",
			SMID:            99,
			DateCreated:     "2021-11-26T06:22:19Z",
			OOFShard:        "1",
			Delivery: entities.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: entities.Payment{
				Transaction:  "test-transaction-3",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1817,
				PaymentDT:    1637907727,
				Bank:         "alpha",
				DeliveryCost: 1500,
				GoodsTotal:   317,
				CustomFee:    0,
			},
			Items: []entities.Item{
				{
					ChrtID:      9934932,
					TrackNumber: "TEST789",
					Price:       453,
					RID:         "ab4219087a764ae0btest",
					Name:        "Mascaras",
					Sale:        30,
					Size:        "0",
					TotalPrice:  317,
					NmID:        2389212,
					Brand:       "Vivienne Sabo",
					Status:      202,
				},
			},
		}
		
		err := repo.SaveOrder(ctx, order)
		assert.NoError(t, err)
		
		err = repo.DeleteOrder(ctx, "test-order-3")
		assert.NoError(t, err)
		
		_, err = repo.GetOrder(ctx, "test-order-3")
		assert.ErrorIs(t, err, ErrOrderNotFound)
	})
	
	t.Run("Clear Orders", func(t *testing.T) {
		ctx := context.Background()
		
		order := entities.Order{
			OrderUID:        "test-order-4",
			TrackNumber:     "TEST101",
			Entry:           "WBIL",
			Locale:          "en",
			InternalSig:     "",
			CustomerID:      "test-customer",
			DeliveryService: "meest",
			ShardKey:        "9",
			SMID:            99,
			DateCreated:     "2021-11-26T06:22:19Z",
			OOFShard:        "1",
			Delivery: entities.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: entities.Payment{
				Transaction:  "test-transaction-4",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1817,
				PaymentDT:    1637907727,
				Bank:         "alpha",
				DeliveryCost: 1500,
				GoodsTotal:   317,
				CustomFee:    0,
			},
			Items: []entities.Item{
				{
					ChrtID:      9934933,
					TrackNumber: "TEST101",
					Price:       453,
					RID:         "ab4219087a764ae0btest",
					Name:        "Mascaras",
					Sale:        30,
					Size:        "0",
					TotalPrice:  317,
					NmID:        2389212,
					Brand:       "Vivienne Sabo",
					Status:      202,
				},
			},
		}
		
		err := repo.SaveOrder(ctx, order)
		assert.NoError(t, err)
		
		err = repo.ClearOrders(ctx)
		assert.NoError(t, err)
		
		orders, err := repo.GetAllOrders(ctx, 100, 0)
		assert.NoError(t, err)
		assert.Empty(t, orders)
	})
}
