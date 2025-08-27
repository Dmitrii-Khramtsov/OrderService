// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/server.go
package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
	repo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/migrations"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/handler"
)

type Shutdownable interface {
	Shutdown(ctx context.Context) error
}

type App struct {
	Server        *http.Server
	Cache         cache.Cache
	Logger        logger.LoggerInterface
	Service       application.OrderServiceInterface
	Handler       *handler.OrderHandler
	Repo          repo.OrderRepository
	KafkaConsumer kafka.ConsumerInterface
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	l, err := newLogger()
	if err != nil {
		return nil, err
	}

	c := newCache(l, cfg.Cache.Capacity)
	db, err := newDatabaseConnection(l)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx := context.Background()
	if err := runMigrations(ctx, db, l); err != nil {
		l.Error("failed to run migrations", zap.Error(err))
		return nil, err
	}

	dsn := os.Getenv("POSTGRES_DSN")
	rp, err := newPostgresOrderRepository(dsn, l)
	if err != nil {
		return nil, err
	}

	svc := newOrderService(c, l, rp, cfg.Cache.GetAllLimit)
	h := newOrderHandler(svc, l)

	brokers := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")

	kc := newKafkaConsumer(brokers, topic, groupID, svc, l)
	r := newRouter(h)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	srv := newHTTPServer(port, r)

	return &App{
		Server:        srv,
		Cache:         c,
		Logger:        l,
		Service:       svc,
		Handler:       h,
		Repo:          rp,
		KafkaConsumer: kc,
	}, nil
}

func newLogger() (logger.LoggerInterface, error) {
	mode := logger.DEV
	if os.Getenv("LOG_MODE") == "production" {
		mode = logger.PROD
	}
	return logger.NewLogger(mode)
}

func newCache(l logger.LoggerInterface, capacity int) cache.Cache {
	return cache.NewOrderLRUCache(l, capacity)
}

func newDatabaseConnection(l logger.LoggerInterface) (*sql.DB, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		l.Error("failed to connect to db", zap.Error(err))
		return nil, repo.ErrDatabaseConnectionFailed
	}
	return db, nil
}

func runMigrations(ctx context.Context, db *sql.DB, l logger.LoggerInterface) error {
	if err := migrations.RunMigrations(ctx, db, l); err != nil {
		l.Error("failed to run migrations", zap.Error(err))
		return err
	}
	l.Info("migrations completed successfully")
	return nil
}

func newPostgresOrderRepository(dsn string, l logger.LoggerInterface) (repo.OrderRepository, error) {
	return repo.NewPostgresOrderRepository(dsn, l)
}

func newOrderService(c cache.Cache, l logger.LoggerInterface, rp repo.OrderRepository, limit int) application.OrderServiceInterface {
	return application.NewOrderService(c, l, rp, limit)
}

func newOrderHandler(svc application.OrderServiceInterface, l logger.LoggerInterface) *handler.OrderHandler {
	return handler.NewOrderHandler(svc, l)
}

func newKafkaConsumer(brokers string, topic, groupID string, svc application.OrderServiceInterface, l logger.LoggerInterface) kafka.ConsumerInterface {
	kc := kafka.NewConsumer([]string{brokers}, topic, groupID, svc, l)
	kc.Start()
	return kc
}

func newRouter(h *handler.OrderHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/orders", h.Create)
	r.Get("/orders/{id}", h.GetByID)
	r.Get("/orders", h.GetAll)
	r.Delete("/orders/{id}", h.Delete)
	r.Delete("/orders", h.Clear)
	fs := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fs)
	return r
}

func newHTTPServer(port string, r *chi.Mux) *http.Server {
	return &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
}

func (a *App) Run() {
	a.Logger.Info("server starting", zap.String("addr", a.Server.Addr))
	go a.RestoreCacheFromDB(context.Background())
	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Error("server listen failed", zap.Error(err))
		}
	}()
}

func (a *App) RestoreCacheFromDB(ctx context.Context) {
	orders, err := a.Service.GetAllFromDB(ctx)
	if err != nil {
		a.Logger.Error("failed to restore cache from DB", zap.Error(err))
		return
	}
	for _, o := range orders {
		a.Cache.Set(o.OrderUID, o)
	}
	a.Logger.Info("cache restored from DB", zap.Int("count", len(orders)))
}

func (a *App) Shutdown(ctx context.Context) {
	a.Logger.Info("shutdown initiated")
	if err := a.KafkaConsumer.Shutdown(ctx); err != nil {
		a.Logger.Error("failed to shutdown kafka consumer", zap.Error(err))
	} else {
		a.Logger.Info("kafka consumer stopped gracefully")
	}
	if err := a.Server.Shutdown(ctx); err != nil {
		a.Logger.Error("server forced to shutdown", zap.Error(err))
	} else {
		a.Logger.Info("server stopped gracefully")
	}
	resources := []struct {
		name string
		res  Shutdownable
	}{
		{"cache", a.Cache},
		{"logger", a.Logger},
		{"repository", a.Repo},
	}
	for _, resource := range resources {
		if err := resource.res.Shutdown(ctx); err != nil {
			a.Logger.Error("failed to shutdown resource",
				zap.String("resource", resource.name),
				zap.Error(err),
			)
		} else {
			a.Logger.Info("resource stopped gracefully",
				zap.String("resource", resource.name),
			)
		}
	}
	a.Logger.Info("shutdown completed")
}
