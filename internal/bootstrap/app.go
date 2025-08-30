// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/app.go
package bootstrap

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/factory"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
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
	Logger        domainrepo.Logger
	Cache         domainrepo.Cache
	CacheRestorer *cache.CacheRestorer
	Repo          domainrepo.OrderRepository
	Service       application.OrderServiceInterface
	Handler       *handler.OrderHandler
	KafkaConsumer domainrepo.EventConsumer
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

	cacheRestorer := factory.NewCacheRestorer(cfg, c, rp, l)

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
