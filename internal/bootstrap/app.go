// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/app.go
package bootstrap

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	repo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/handler"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/router"
)

type Shutdownable interface {
	Shutdown(ctx context.Context) error
}

type DBWrapper struct {
	DB *sqlx.DB
}

func (dw *DBWrapper) Shutdown(ctx context.Context) error {
	return dw.DB.Close()
}

type App struct {
	Server        *http.Server
	Logger        logger.LoggerInterface
	Cache         cache.Cache
	CacheRestorer *cache.CacheRestorer
	Repo          repo.OrderRepository
	Service       application.OrderServiceInterface
	Handler       *handler.OrderHandler
	KafkaConsumer kafka.ConsumerInterface
	DB            Shutdownable
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		return nil, err
	}

	l, err := factory.NewLogger(cfg)
	if err != nil {
		return nil, err
	}
	c := factory.NewCache(l, cfg.Cache.Capacity)
	db, err := factory.NewDatabase(cfg, l)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(context.Background(), db, cfg.Migrations.MigrationsPath, l, cfg); err != nil {
		return nil, err
	}

	rp, err := factory.NewOrderRepository(cfg, db, l)
	if err != nil {
		return nil, err
	}

	cacheRestorer := factory.NewCacheRestorer(c, rp, l)

	svc := application.NewOrderService(c, l, rp, cfg.Cache.GetAllLimit)

	h := handler.NewOrderHandler(svc, l)
	r := router.New(h)
	srv := factory.NewHTTPServer(cfg.Server.Port, r)

	kc := factory.NewKafkaConsumer(cfg.Kafka, svc, l)

	return &App{
		Server:        srv,
		Logger:        l,
		Cache:         c,
		CacheRestorer: cacheRestorer,
		Repo:          rp,
		Service:       svc,
		Handler:       h,
		KafkaConsumer: kc,
		DB:            &DBWrapper{DB: db},
	}, nil
}
